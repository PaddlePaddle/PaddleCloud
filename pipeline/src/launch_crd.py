# Copyright 2019 kubeflow.org.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import datetime
import logging
import time

from kubernetes import client as k8s_client
from kubernetes.client import rest

logger = logging.getLogger(__name__)


class K8sCR(object):
    def __init__(self, group, plural, version, client):
        self.group = group
        self.plural = plural
        self.version = version
        self.client = k8s_client.CustomObjectsApi(client)

    def wait_for_condition(self,
                           namespace,
                           name,
                           expected_conditions=None,
                           error_phases=None,
                           timeout=datetime.timedelta(days=365),
                           polling_interval=datetime.timedelta(seconds=10),
                           status_callback=None,
                           wait_created=False):
        """Waits until any of the specified conditions occur.
            Args:
              namespace: namespace for the CR.
              name: Name of the CR.
              expected_conditions: A list of conditions. Function waits until any of the
                supplied conditions is reached.
              error_phases: A list of phase string, if the phase of CR in this list, means
                some error occurs.
              timeout: How long to wait for the CR.
              polling_interval: How often to poll for the status of the CR.
              status_callback: (Optional): Callable. If supplied this callable is
                invoked after we poll the CR. Callable takes a single argument which is the CR.
              wait_created: wait until the object has been created.
        """
        if not expected_conditions and not error_phases:
            logger.info("both expected_conditions and error_phases not be set")
            return

        if expected_conditions is None:
            expected_conditions = []
        if error_phases is None:
            error_phases = []

        end_time = datetime.datetime.now() + timeout
        while True:
            try:
                results = self.client.get_namespaced_custom_object(
                    self.group, self.version, namespace, self.plural, name)
            except rest.ApiException as err:
                if str(err.status) == "404" and wait_created:
                    print("Waiting for for %s/%s %s in namespace %s been created",
                          self.group, self.plural, name, namespace)
                    time.sleep(polling_interval.seconds)
                    continue

                logger.error("There was a problem waiting for %s/%s %s in namespace %s; Exception: %s",
                             self.group, self.plural, name, namespace, err)
                raise

            if results:
                if status_callback:
                    status_callback(results)
                expected, condition = self.is_expected_conditions(results, expected_conditions)
                if expected:
                    logger.info("%s/%s %s in namespace %s has reached the expected condition: %s.",
                                self.group, self.plural, name, namespace, condition)
                    return results
                else:
                    if condition:
                        logger.info("Current condition of %s/%s %s in namespace %s is %s.",
                                    self.group, self.plural, name, namespace, condition)

                # phase = results.get("status", {}).get("phase", "")
                errored, _ = self.is_expected_conditions(results, error_phases)
                if errored:
                    raise Exception("There are some errors in {0}/{1} {2} in namespace {3}, phase {4}."
                                    .format(self.group, self.plural, name, namespace, condition))

            if datetime.datetime.now() + polling_interval > end_time:
                raise Exception(
                    "Timeout waiting for {0}/{1} {2} in namespace {3} to enter one of the "
                    "conditions {4}.".format(self.group, self.plural, name, namespace, expected_conditions))

            time.sleep(polling_interval.seconds)

    def is_expected_conditions(self, inst, expected_conditions):
        conditions = inst.get('status', {}).get("conditions")
        if not conditions:
            return False, ""
        if conditions[-1]["type"] in expected_conditions and conditions[-1]["status"] == "True":
            return True, conditions[-1]["type"]
        else:
            return False, conditions[-1]["type"]

    def patch(self, spec):
        """Apply custom resource
          Args:
            spec: The spec for the CR
        """
        name = spec["metadata"]["name"]
        namespace = spec["metadata"].get("namespace", "default")
        logger.info("Patching %s/%s %s in namespace %s.",
                    self.group, self.plural, name, namespace)
        api_response = self.client.patch_namespaced_custom_object(
          self.group, self.version, namespace, self.plural, name, spec)
        logger.info("Patched %s/%s %s in namespace %s.",
                    self.group, self.plural, name, namespace)
        return api_response

    def create(self, spec):
        """Create a CR.
        Args:
          spec: The spec for the CR.
        """
        # Create a Resource
        namespace = spec["metadata"].get("namespace", "default")
        logger.info("Creating %s/%s %s in namespace %s.",
                    self.group, self.plural, spec["metadata"]["name"], namespace)
        api_response = self.client.create_namespaced_custom_object(
            self.group, self.version, namespace, self.plural, spec)
        logger.info("Created %s/%s %s in namespace %s.",
                    self.group, self.plural, spec["metadata"]["name"], namespace)
        return api_response

    def apply(self, spec):
        """Create or update a CR
        Args:
          spec: The spec for the CR.
        """
        name = spec["metadata"]["name"]
        namespace = spec["metadata"].get("namespace", "default")
        logger.info("Apply %s/%s %s in namespace %s.",
                    self.group, self.plural, name, namespace)

        try:
            api_response = self.client.create_namespaced_custom_object(
                self.group, self.version, namespace, self.plural, spec)
            return api_response
        except rest.ApiException as err:
            if str(err.status) != "409":
                raise

        logger.info("Already exists now begin updating")
        api_response = self.client.patch_namespaced_custom_object(
            self.group, self.version, namespace, self.plural, name, spec)
        logger.info("Applied %s/%s %s in namespace %s.",
                    self.group, self.plural, name, namespace)
        return api_response

    def delete(self, name, namespace):
        logger.info("Deleting %s/%s %s in namespace %s.",
                    self.group, self.plural, name, namespace)
        api_response = self.client.delete_namespaced_custom_object(
            self.group, self.version, namespace, self.plural,
            name, propagation_policy="Foreground")
        logger.info("Deleted %s/%s %s in namespace %s.",
                    self.group, self.plural, name, namespace)
        return api_response

    def run(self, spec, action):
        inst_spec = self.get_spec(spec)
        name = inst_spec.get("metadata").get("name")
        namespace = inst_spec.get("metadata").get("namespace")
        print(f"The spec of crd is {inst_spec}")
        if action == "create":
            response = self.create(inst_spec)
        elif action == "patch":
            response = self.patch(inst_spec)
        elif action == "apply":
            response = self.apply(inst_spec)
        elif action == "delete":
            response = self.delete(name, namespace)
            print(f"Delete {self.group}/{self.version}/{self.plural} have response {response}")
            return
        else:
            raise Exception("action must be one of create/patch/apply/delete")

        print(f"{action} {self.group}/{self.version}/{self.plural} have response {response}")

        expected_conditions, error_phases = self.get_action_status(action=action)

        self.wait_for_condition(namespace, name,
                                expected_conditions=expected_conditions,
                                error_phases=error_phases,)

    def get_spec(self, spec):
        raise NotImplemented("This method must be implemented")

    def get_action_status(self, action=None):
        raise NotImplemented("This method must be implemented")
