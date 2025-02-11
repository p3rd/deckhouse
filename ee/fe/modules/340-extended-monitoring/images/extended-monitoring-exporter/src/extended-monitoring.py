#!/usr/bin/env python3

# Copyright 2021 Flant JSC
# Licensed under the Deckhouse Platform Enterprise Edition (EE) license.
# See https://github.com/deckhouse/deckhouse/blob/main/ee/LICENSE
from concurrent.futures.thread import ThreadPoolExecutor
from itertools import chain
from threading import Thread
from time import sleep
import logging

import kubernetes
import copy
import sys

from abc import ABC, abstractmethod
from http.server import BaseHTTPRequestHandler, HTTPServer
from socketserver import ThreadingMixIn

kubernetes.config.load_incluster_config()

logging.basicConfig(format='[%(asctime)s] - %(message)s', level=logging.INFO)

EXTENDED_MONITORING_ANNOTATION_THRESHOLD_PREFIX = "threshold.extended-monitoring.flant.com/"
EXTENDED_MONITORING_ENABLED_ANNOTATION = "extended-monitoring.flant.com/enabled"

DEFAULT_SERVER_ADDRESS = '0.0.0.0'
DEFAULT_PORT = 8080


class ThreadingHTTPServer(ThreadingMixIn, HTTPServer):
    daemon_threads = True


class Annotated(ABC):
    default_thresholds = {}

    def __init__(self, namespace, name, kube_annotations):
        self.namespace = namespace
        self.name = name
        self.enabled = True

        if kube_annotations:
            if not {EXTENDED_MONITORING_ENABLED_ANNOTATION: "false"}.items() <= kube_annotations.items():
                self.thresholds = copy.deepcopy(self.default_thresholds)
                for name, value in kube_annotations.items():
                    if name.startswith(EXTENDED_MONITORING_ANNOTATION_THRESHOLD_PREFIX):
                        self.thresholds.update(
                            {name.replace(EXTENDED_MONITORING_ANNOTATION_THRESHOLD_PREFIX, ""): value})
            else:
                self.enabled = False
        else:
            self.thresholds = copy.deepcopy(self.default_thresholds)

    @classmethod
    def list_threshold_annotated_objects(cls, namespace):
        for kube_object in cls.list(namespace):
            yield cls(namespace, kube_object.metadata.name, kube_object.metadata.annotations)

    @property
    def formatted(self):
        to_return = ""

        if self.enabled:
            to_return += 'extended_monitoring_{}_enabled{{namespace="{}", {}="{}"}} {}\n'.format(
                self.kind.lower(),
                self.namespace,
                self.kind.lower(),
                self.name, 1)

            if hasattr(self, "thresholds"):
                for k, v in self.thresholds.items():
                    to_return += 'extended_monitoring_{}_threshold{{namespace="{}", threshold="{}", {}="{}"}} {}\n'.format(
                        self.kind.lower(),
                        self.namespace,
                        k,
                        self.kind.lower(),
                        self.name, int(v))
        else:
            to_return += 'extended_monitoring_{}_enabled{{namespace="{}", {}="{}"}} {}\n'.format(
                self.kind.lower(),
                self.namespace,
                self.kind.lower(),
                self.name, 0)

        return to_return

    @property
    @abstractmethod
    def kind(self):
        pass

    @property
    @abstractmethod
    def api(self):
        pass

    @classmethod
    @abstractmethod
    def list(cls, namespace):
        pass


class AnnotatedDeployment(Annotated):
    kind = "Deployment"
    api = kubernetes.client.AppsV1Api()

    @classmethod
    def list(cls, namespace):
        return cls.api.list_namespaced_deployment(namespace).items

    default_thresholds = {
        "replicas-not-ready": 0
    }


class AnnotatedStatefulSet(Annotated):
    kind = "StatefulSet"
    api = kubernetes.client.AppsV1Api()

    @classmethod
    def list(cls, namespace):
        return cls.api.list_namespaced_stateful_set(namespace).items

    default_thresholds = {
        "replicas-not-ready": 0
    }


class AnnotatedDaemonSet(Annotated):
    kind = "DaemonSet"
    api = kubernetes.client.AppsV1Api()

    @classmethod
    def list(cls, namespace):
        return cls.api.list_namespaced_daemon_set(namespace).items

    default_thresholds = {
        "replicas-not-ready": 0
    }


class AnnotatedPod(Annotated):
    kind = "Pod"
    api = kubernetes.client.CoreV1Api()

    @classmethod
    def list(cls, namespace):
        return cls.api.list_namespaced_pod(namespace).items

    default_thresholds = {
        "disk-bytes-warning": 85,
        "disk-bytes-critical": 95,
        "disk-inodes-warning": 85,
        "disk-inodes-critical": 90,
        "container-throttling-warning": 25,
        "container-throttling-critical": 50,
    }


class AnnotatedIngress(Annotated):
    kind = "Ingress"
    api = kubernetes.client.ExtensionsV1beta1Api()

    @classmethod
    def list(cls, namespace):
        return cls.api.list_namespaced_ingress(namespace).items

    default_thresholds = {
        "5xx-warning": 10,
        "5xx-critical": 20
    }


class AnnotatedNode(Annotated):
    kind = "Node"
    api = kubernetes.client.CoreV1Api()

    @classmethod
    def list(cls, namespace=None):
        return cls.api.list_node().items

    default_thresholds = {
        "disk-bytes-warning": 70,
        "disk-bytes-critical": 80,
        "disk-inodes-warning": 85,
        "disk-inodes-critical": 90,
        "load-average-per-core-warning": 3,
        "load-average-per-core-critical": 10,
    }


class AnnotatedCronJob(Annotated):
    kind = "CronJob"
    api = kubernetes.client.BatchV1beta1Api()

    @classmethod
    def list(cls, namespace):
        return cls.api.list_namespaced_cron_job(namespace).items


KUBERNETES_OBJECTS = (
  AnnotatedNode,
)
KUBERNETES_NAMESPACED_OBJECTS = (
    AnnotatedDeployment,
    AnnotatedStatefulSet,
    AnnotatedDaemonSet,
    AnnotatedPod,
    AnnotatedIngress,
    AnnotatedCronJob,
)

corev1 = kubernetes.client.CoreV1Api()
apis = kubernetes.client.ApisApi()


def _list_objects(executor, objects, namespace):
    yield from chain.from_iterable(executor.map(lambda k: k.list_threshold_annotated_objects(namespace), objects))


def _get_metrics():
    enabled_nses = []
    quantity = 0

    # iterate over namespaced objects in explicitly enabled via annotation Namespaces
    ns_list = (
        ns.metadata.name for ns in corev1.list_namespace().items
        if ns.metadata.annotations
        and EXTENDED_MONITORING_ENABLED_ANNOTATION in ns.metadata.annotations.keys()
    )

    response = """# HELP extended_monitoring_annotations Extended monitoring annotations
      # TYPE extended_monitoring_annotations gauge\n"""

    with ThreadPoolExecutor(max_workers=8) as executor:
        def _list_in_ns(ns):
            enabled_nses.append('\nextended_monitoring_enabled{{namespace="{}"}} 1'.format(ns))
            yield from _list_objects(executor, KUBERNETES_NAMESPACED_OBJECTS, ns)

        for annotated_object in chain.from_iterable(executor.map(_list_in_ns, ns_list)):
            response += annotated_object.formatted
            quantity += 1

        for annotated_object in _list_objects(executor, KUBERNETES_OBJECTS, None):
            response += annotated_object.formatted
            quantity += 1

    response += '\n'.join(enabled_nses)
    quantity += len(enabled_nses)
    return response, quantity


class GetHandler(BaseHTTPRequestHandler):
    _response = ""

    @classmethod
    def get_metrics(cls):
        # setting string variable is atomic in Python
        cls._response, quantity = _get_metrics()
        logging.info('Metrics are collected successfully. Batches quantity: {}'.format(quantity))

    @classmethod
    def loop_get_metrics(cls):
        while 1:
            try:
                cls.get_metrics()
            except Exception as loop_err:
                logging.info(str(loop_err))
            sleep(30)

    def do_GET(self):
        if self.path == "/ready":
            apis.get_api_versions()
            self.send_response(200)
            self.end_headers()
            return

        if self.path == "/healthz":
            self.send_response(200)
            self.end_headers()
            return

        if self.path == "/metrics":
            self.send_response(200)
            self.send_header('Content-Type',
                             'text/plain; charset=utf-8')
            self.end_headers()
            self.wfile.write(self.__class__._response.encode(encoding="utf-8"))
            return

        self.send_response(404)
        self.end_headers()


if __name__ == '__main__':
    server_address = DEFAULT_SERVER_ADDRESS
    server_port = DEFAULT_PORT

    # Parse host and port
    if len(sys.argv) >= 2:
        server_address = sys.argv[1]
    if len(sys.argv) == 3:
        server_port = int(sys.argv[2])

    # Get metrics once synchronously before starting web server
    GetHandler.get_metrics()
    server = ThreadingHTTPServer((server_address, server_port), GetHandler)

    try:
        # Run metrics renew in background (daemon thread is canceled on the script exit)
        Thread(target=GetHandler.loop_get_metrics, daemon=True).start()

        logging.info('Starting server')
        server.serve_forever()
    except Exception as err:
        logging.info('Shutting down server')
        raise err
