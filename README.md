# VMware vSphere event source 

This is a [Knative eventing](https://knative.dev/docs/eventing/) source which provides a way to receive and forward events from VMware vSphere to Kubernetes platform.

## Installation

vSphere event source usage assumes that you have [Knative](https://knative.dev/docs/install/) platform installed on your Kubernetes cluster. To simplify CRD development, tests and installation we recommend to use [ko](https://github.com/google/ko) command line tool.


After you cloned this repository `VSphereSource` custom resource definition can be installed by running following command:

```
ko apply -f config/
```

Check CRD controller state:

```
kubectl -n knative-sources get statefulsets vsphere-controller-manager
```

## Usage

Once VSphereSource CRD was installed on cluster, in order to start receiving vSphere events you should create k8s secret with vSphere (vCenter or ESXi) account password:

```
kubectl create secret generic vsphere-credentials --from-literal=password=<PASSWORD>
```

If you don't have any event receivers, you may install and use `event-display` service to dump all incoming messages into a log:

```
kubectl apply -f https://github.com/knative/eventing-contrib/blob/master/config/tools/event-display/event-display.yaml
```

And the last step is to create `VSphereSource` object:

```
cat <<EOF | kubectl apply -f -
apiVersion: sources.eventing.triggermesh.dev/v1alpha1
kind: VSphereSource
metadata:
  name: vsphere-sample-source
spec:
  vsphereCredsSecret:
    name: vsphere-credentials
    key: password
  vsphereUser: administrator@vsphere.local
  vsphereUrl: 1.2.3.4
  sink:
    apiVersion: serving.knative.dev/v1alpha1
    kind: Service
    name: event-display
EOF
```
where `vsphereUser` and `vsphereUrl` should be replaced with vCenter or ESXi valid username and address. 
After the object is created you can list pods and see that there are two new are being created - one for the event source and another for event-display service. vSphere events should be available in event-display service pod logs:

```
kubectl logs event-display-qdhb8-deployment-57df8d97b8-fx7zw -c user-container

☁️  cloudevents.Event
Validation: valid
Context Attributes,
  specversion: 0.2
  type: vmware.vsphere.message
  source: 1.2.3.4
  id: 2824
  time: 2019-09-02T10:49:38.112999Z
  contenttype: application/json
Data,
  "Alarm 'Host hardware temperature status' on ns3135023.ip-1-2-3.eu changed from Gray to Green"

☁️  cloudevents.Event
Validation: valid
Context Attributes,
  specversion: 0.2
  type: vmware.vsphere.message
  source: 1.2.3.4
  id: 2971
  time: 2019-09-02T15:58:43.585849Z
  contenttype: application/json
Data,
  "foo-vm on ns3135023.ip-1-2-3.eu in datacenter-1 is powered off"

☁️  cloudevents.Event
Validation: valid
Context Attributes,
  specversion: 0.2
  type: vmware.vsphere.message
  source: 1.2.3.4
  id: 2973
  time: 2019-09-02T15:58:59.054846Z
  contenttype: application/json
Data,
  "Removed foo-vm on ns3135023.ip-1-2-3.eu from datacenter-1"

```

## Support

We would love your feedback and help on these sources, so don't hesitate to let us know what is wrong and how we could improve them, just file an [issue](https://github.com/triggermesh/vsphere-source/issues/new) or join those of use who are maintaining them and submit a [PR](https://github.com/triggermesh/vsphere-source/compare).

## Code of conduct

This plugin is by no means part of [CNCF](https://www.cncf.io/) but we abide by its [code of conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).


