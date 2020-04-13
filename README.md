# compute-operator

## Pre Req:
- OCP 4 installed

#### Clone it

    git clone https://github.com/luis5tb/worker-osp-operator
    cd worker-osp-operator

#### Create the operator

This is optional, a prebuild operator from quay.io/ltomasbo/compute-operator could be used, e.g. quay.io/ltomasbo/compute-operator:v0.0.1 .

Build the image, using your custom registry you have write access to

    operator-sdk build <image e.g quay.io/ltomasbo/compute-operator:v0.0.X>

Replace `image:` in deploy/operator.yaml with your custom registry

    sed -i 's|REPLACE_IMAGE|quay.io/ltomasbo/compute-operator:v0.0.X|g' deploy/operator.yaml
    podman push --authfile ~/ltomasbo-auth.json quay.io/ltomasbo/compute-operator:v0.0.X

#### Install the operator

Create CRDs
    
    oc create -f deploy/crds/compute.openstack.org_computeopenstacks_crd.yaml

Create role, role_binding and service_account

    oc create -f deploy/role.yaml
    oc create -f deploy/role_binding.yaml
    oc create -f deploy/service_account.yaml

Install the operator

    oc create -f deploy/operator.yaml

If necessary check logs with

    POD=`oc get pods -l name=worker-osp-operator --field-selector=status.phase=Running -o name | head -1 -`; echo $POD
    oc logs $POD -f

Create custom resource for a compute node which specifies the needed information (e.g.: `deploy/crds/compute.openstack.org_v1alpha1_computeopenstack_cr.yaml`):

    apiVersion: compute.openstack.org/v1alpha1
    kind: ComputeOpenStack
    metadata:
      name: example-computeopenstack
    spec:
      # Add fields here
      roleName: worker-osp
      clusterName: ostest
      baseWorkerMachineSetName: ostest-worker-0
      k8sServiceIp: 172.30.0.1
      apiIntIp: 192.168.111.5
      workers: 1
      corePinning: "4-7"   # Optional

Apply the CR:

    oc apply -f deploy/crds/compute.openstack.org_v1alpha1_computeopenstack_cr.yaml
    
    oc get pods -n openshift-machine-api
    NAME                                 READY   STATUS    RESTARTS   AGE
    worker-osp-operator-ffd64796-vshg6   1/1     Running   0          119s

Get the generated machineconfig and machinesets
    oc get machineset  -n openshift-machine-api
    oc get machineconfigpool
    oc get machineconfig


## POST steps to add compute workers

Edit the computeopenstack CR:

    oc -n openshift-machine-api edit computeopenstacks.compute.openstack.org example-computeopenstack
    # Modify the number of workers and exit

    oc get machineset -n openshift-machine-api
    # check the desired amount has been updated

## Cleanup

First delete all instances running on the OCP:

    oc delete -f deploy/crds/compute.openstack.org_v1alpha1_computeopenstack_cr.yaml
    oc delete -f deploy/operator.yaml
    oc delete -f deploy/role.yaml
    oc delete -f deploy/role_binding.yaml
    oc delete -f deploy/service_account.yaml
    oc delete -f deploy/crds/compute.openstack.org_computeopenstacks_crd.yaml