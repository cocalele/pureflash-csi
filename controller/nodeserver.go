/*
Copyright 2023 LiuLele

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"context"

	"golang.org/x/sys/unix"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

const topologyKeyNode = "topology.controller.csi/node"

type nodeServer struct {
	nodeID            string
	ephemeral         bool
	maxVolumesPerNode int64

}

func newNodeServer(nodeID string, ephemeral bool, maxVolumesPerNode int64) *nodeServer {

	// revive existing volumes at start of node server

	return &nodeServer{
		nodeID:            nodeID,
		ephemeral:         ephemeral,
		maxVolumesPerNode: maxVolumesPerNode,


	}
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {

	// Check arguments
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	targetPath := req.GetTargetPath()

	if req.GetVolumeCapability().GetBlock() != nil &&
		req.GetVolumeCapability().GetMount() != nil {
		return nil, status.Error(codes.InvalidArgument, "cannot have both block and mount access type")
	}

	var accessTypeMount, accessTypeBlock bool
	cap := req.GetVolumeCapability()

	if cap.GetBlock() != nil {
		accessTypeBlock = true
	}
	if cap.GetMount() != nil {
		accessTypeMount = true
	}

	// sanity checks (probably more sanity checks are needed later)
	if accessTypeBlock && accessTypeMount {
		return nil, status.Error(codes.InvalidArgument, "cannot have both block and mount access type")
	}


	if req.GetVolumeCapability().GetBlock() != nil {

		output, err := bindMountPfbd(req., targetPath)
		if err != nil {
			return nil, fmt.Errorf("unable to bind mount lv: %w output:%s", err, output)
		}
		// FIXME: VolumeCapability is a struct and not the size
		klog.Infof("block lv %s size:%s created at:%s", req.GetVolumeId(), req.GetVolumeCapability(),  targetPath)

	} else if req.GetVolumeCapability().GetMount() != nil {

		output, err := mountPfbd(req.VolumeId, targetPath, req.GetVolumeCapability().GetMount().GetFsType())
		if err != nil {
			return nil, fmt.Errorf("unable to mount lv: %w output:%s", err, output)
		}
		// FIXME: VolumeCapability is a struct and not the size
		klog.Infof("mounted lv %s size:%s created at:%s", req.GetVolumeId(), req.GetVolumeCapability(), targetPath)

	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {

	// TODO
	// implement deletion of ephemeral volumes
	volID := req.GetVolumeId()

	klog.Infof("NodeUnpublishRequest: %s", req)
	// Check arguments
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	umountPfbd(req.GetTargetPath())

	// ephemeral volumes start with "csi-"
	if strings.HasPrefix(volID, "csi-") {
		// remove ephemeral volume here
		deletePfbdVolume(volID)

	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capability missing in request")
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {

	topology := &csi.Topology{
		Segments: map[string]string{topologyKeyNode: ns.nodeID},
	}

	return &csi.NodeGetInfoResponse{
		NodeId:             ns.nodeID,
		MaxVolumesPerNode:  ns.maxVolumesPerNode,
		AccessibleTopology: topology,
	}, nil
}

func (ns *nodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
					},
				},
			},
		},
	}, nil
}

func (ns *nodeServer) NodeGetVolumeStats(ctx context.Context, in *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {

	var fs unix.Statfs_t

	err := unix.Statfs(in.GetVolumePath(), &fs)
	if err != nil {
		return nil, err
	}

	diskFree := int64(fs.Bfree) * int64(fs.Bsize)
	diskTotal := int64(fs.Blocks) * int64(fs.Bsize)

	inodesFree := int64(fs.Ffree)
	inodesTotal := int64(fs.Files)

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Available: diskFree,
				Total:     diskTotal,
				Used:      diskTotal - diskFree,
				Unit:      csi.VolumeUsage_BYTES,
			},
			{
				Available: inodesFree,
				Total:     inodesTotal,
				Used:      inodesTotal - inodesFree,
				Unit:      csi.VolumeUsage_INODES,
			},
		},
	}, nil
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {

	// Check arguments
	if req.GetCapacityRange() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	capacity := int64(req.GetCapacityRange().GetRequiredBytes())

	volID := req.GetVolumeId()
	volPath := req.GetVolumePath()
	if len(volPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume path not provided")
	}

	info, err := os.Stat(volPath)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Could not get file information from %s: %v", volPath, err)
	}

	isBlock := false
	m := info.Mode()
	if !m.IsDir() {
		klog.Warning("volume expand request on block device: filesystem resize has to be done externally")
		isBlock = true
	}

	//output, err := extendLVS(ns.vgName, volID, uint64(capacity), isBlock)
	klog.Error("NodeExpandVolume not implemented");
	_ = volID
	_ =  isBlock



	return &csi.NodeExpandVolumeResponse{
		CapacityBytes: capacity,
	}, nil

}

func getPfbdDevName(volumeName string) (string,error) {

	cmdStr := "set -o pipefail; pfkd_helper -l | grep '%s' | awk '{print $1}'"
	if devName, err := exec.Command("bash", "-c", cmdStr).Output(); err != nil {
		return "", fmt.Errorf("volume:%s not attached", volumeName)
	} else {
		return string(devName), nil
	}
}
func attachPfbd(volumeName string) (string, error){
	var devPath string
	if devName, err := getPfbdDevName(volumeName); err != nil {
		klog.Infof("Not found volume:%s, try to attach it", volumeName)
		if _, err := exec.Command("bash", "-c", "pfkd_helper -a "+volumeName).Output();err != nil {
			klog.Errorf("Failed to attach volume:%s", volumeName)
			return "", fmt.Errorf("failed to attach volume:%s", volumeName)
		}

		//let's try again
		if devName, err := getPfbdDevName(volumeName); err != nil {
			klog.Infof("Still can't found volume:%s", volumeName)
			return "", fmt.Errorf("still can't found volume:%s", volumeName)
		}else {
			klog.Infof("volume:%s device name:%s \n", volumeName, devName)
			devPath = "/dev/" + string(devName)
		}
	} else {
		klog.Infof("INFO volume:%s device name:%s \n", volumeName, devName)
		devPath = "/dev/" + string(devName)
	}

	return devPath, nil
}
func mountPfbd(volumeName, mountPath string, fsType string) (string, error) {


	devPath, err := attachPfbd(volumeName)
	if err != nil {
		return "", err
	}
	formatted := false
	forceFormat := false
	if fsType == "" {
		fsType = "ext4"
	}
	// check for already formatted
	cmd := exec.Command("blkid", devPath)
	out, err2 := cmd.CombinedOutput()
	if err2 != nil {
		klog.Infof("unable to check if %s is already formatted:%v", devPath, err2)
	}
	matches := fsTypeRegexp.FindStringSubmatch(string(out))
	if len(matches) > 1 {
		if matches[1] == "xfs_external_log" { // If old xfs signature was found
			forceFormat = true
		} else {
			if matches[1] != fsType {
				return string(out), fmt.Errorf("target fsType is %s but %s found", fsType, matches[1])
			}

			formatted = true
		}
	}

	if !formatted {
		formatArgs := []string{}
		if forceFormat {
			formatArgs = append(formatArgs, "-f")
		}
		formatArgs = append(formatArgs, devPath)

		klog.Infof("formatting with mkfs.%s %s", fsType, strings.Join(formatArgs, " "))
		cmd = exec.Command(fmt.Sprintf("mkfs.%s", fsType), formatArgs...) //nolint:gosec
		out, err = cmd.CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("unable to format lv:%s err:%w", devPath, err)
		}
	}

	err = os.MkdirAll(mountPath, 0777|os.ModeSetgid)
	if err != nil {
		return string(out), fmt.Errorf("unable to create mount directory for lv:%s err:%w", devPath, err)
	}

	// --make-shared is required that this mount is visible outside this container.
	mountArgs := []string{"--make-shared", "-t", fsType, devPath, mountPath}
	klog.Infof("mount volume command: mount %s", mountArgs)
	cmd = exec.Command("mount", mountArgs...)
	out, err = cmd.CombinedOutput()
	if err != nil {
		mountOutput := string(out)
		if !strings.Contains(mountOutput, "already mounted") {
			return string(out), fmt.Errorf("unable to mount %s to %s err:%w output:%s", devPath, mountPath, err, out)
		}
	}
	err = os.Chmod(mountPath, 0777|os.ModeSetgid)
	if err != nil {
		return "", fmt.Errorf("unable to change permissions of volume mount %s err:%w", mountPath, err)
	}
	klog.Infof("mount output:%s", out)
	return "", nil
}


func detachPfbd(volumeName string) error {


	cmdStr := fmt.Sprintf("pfkd_helper -d %s", volumeName)
	if _, err := exec.Command("bash", "-c", cmdStr).Output(); err != nil {
		klog.Errorf("Failed to detach volume:%s error:%s", volumeName, err.Error())
		return err
	}
	return nil
}

func umountPfbd(targetPath string) {
	cmd := exec.Command("umount", "--lazy", "--force", targetPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("unable to umount %s output:%s err:%w", targetPath, string(out), err)
	}
}


func bindMountPfbd(devPath, mountPath string) (string, error) {

	_, err := os.Create(mountPath)
	if err != nil {
		return "", fmt.Errorf("unable to create mount directory for lv:%s err:%w", devPath, err)
	}

	// --make-shared is required that this mount is visible outside this container.
	// --bind is required for raw block volumes to make them visible inside the pod.
	mountArgs := []string{"--make-shared", "--bind", devPath, mountPath}
	klog.Infof("bindmountlv command: mount %s", mountArgs)
	cmd := exec.Command("mount", mountArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		mountOutput := string(out)
		if !strings.Contains(mountOutput, "already mounted") {
			return string(out), fmt.Errorf("unable to mount %s to %s err:%w output:%s", devPath, mountPath, err, out)
		}
	}
	err = os.Chmod(mountPath, 0777|os.ModeSetgid)
	if err != nil {
		return "", fmt.Errorf("unable to change permissions of volume mount %s err:%w", mountPath, err)
	}
	klog.Infof("bindmountlv output:%s", out)
	return "", nil
}

func deletePfbdVolume(volName string){
	klog.Errorf("deletePfbdVolume not impl")
}