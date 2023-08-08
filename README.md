# PureFlash K8S CSI pulgin

PureFlash is an open-source storage system, it provide block storage service for cloud and legacy applications. PureFlash has a very high performance that can provides 1 Million IOPS random write per node.
This project is a plugin to:
  - deploy PureFlash in k8s with yaml
  - CSI plugin to use PureFlash to provide PV for k8s applications
  
## Prequisite
get this project from github or gitee:
```
	# git clone https://gitee.com/cocalele/pureflash-csi.git
	or
	# git clone https://github.com/cocalele/pureflash-csi.git
```

## Deploy storage service
This will deploy the PureFlash cluster in your k8s. Like other cloud native storage service, e.g. openEBS, PureFlash convert the physical disks to a virtual storage service, with features like thin-provision, snapshot, replication, HA, etc. 

1. modify the value of `PFS_DISKS` in file `pfs-cluster/deploy/pfs.yaml `, change it to your real NVMe device name.
```
          - name: PFS_DISKS
            value: "/dev/nvme1n1,/dev/nvme2n1"
```
The physical disks should be clean and has no data on it. It will be managed by PureFlash.

2. apply the following yaml files:
```
	# kubectl apply -f pfs-cluster/deploy/namespace.yaml
	# kubectl apply -f pfs-cluster/deploy/pfzk.yaml
	# kubectl apply -f pfs-cluster/deploy/pfdb.yaml
	# kubectl apply -f pfs-cluster/deploy/pfc.yaml
	# kubectl apply -f pfs-cluster/deploy/pfs.yaml
```

Please make sure all pods have running correctly before execute next command.

## Deploy CSI plugin
apply the following yaml files:
```
	# kubectl apply -f deploy/rbac-csi-pfbd.yaml
	# kubectl apply -f deploy/csi-pfbd-driverinfo.yaml
	# kubectl apply -f deploy/csi-pfbd-controller.yaml
	# kubectl apply -f deploy/csi-pfbd-node.yaml
```
PureFlash CSI use native kernel driver as client interface. A kernel driver represent volume as standard block device in Linux so it is  compatible with all applications. Also it can achieve higher performance than iSCSI.

Kernel drivers can be found under `modules/<kernel_version>` directory. more kernel version modules is continuously adding here.
run command `insmod pfkd.ko` before you creating PV.

## Create storage class to use PureFlash
There's an exmaple in this project, 
```
# cat examples/sc.yaml 
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: pfbd-rep2
provisioner: pfbd.csi.pureflash
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
parameters:
  replica_count: "2"

```
`kubectl apply -f  examples/sc.yaml` will create a SC of 2 replica volume.
`kubectl apply -f  examples/my-busybox.yaml` will use this SC to create PVC.

