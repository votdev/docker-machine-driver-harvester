package builder

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	kubevirtv1 "kubevirt.io/client-go/api/v1"

	clientv1 "github.com/harvester/go-harvester/pkg/client/generated/v1"
)

const (
	defaultVMGenerateName = "harv-"
	defaultVMNamespace    = "default"

	defaultVMCPUCores = 1
	defaultVMMemory   = "256Mi"
)

type VMBuilder struct {
	vm              *clientv1.VirtualMachine
	sshNames        []string
	dataVolumeNames []string
	nicNames        []string
}

func NewVMBuilder(creator string) *VMBuilder {
	vmLabels := map[string]string{
		"harvester.cattle.io/creator": creator,
	}
	objectMeta := metav1.ObjectMeta{
		Namespace:    defaultVMNamespace,
		GenerateName: defaultVMGenerateName,
		Labels:       vmLabels,
	}
	running := pointer.BoolPtr(false)
	cpu := &kubevirtv1.CPU{
		Cores: defaultVMCPUCores,
	}
	resources := kubevirtv1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(defaultVMMemory),
		},
	}
	template := &kubevirtv1.VirtualMachineInstanceTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: vmLabels,
		},
		Spec: kubevirtv1.VirtualMachineInstanceSpec{
			Domain: kubevirtv1.DomainSpec{
				CPU: cpu,
				Devices: kubevirtv1.Devices{
					Disks:      []kubevirtv1.Disk{},
					Interfaces: []kubevirtv1.Interface{},
				},
				Resources: resources,
			},
			Networks: []kubevirtv1.Network{},
			Volumes:  []kubevirtv1.Volume{},
		},
	}

	vm := &clientv1.VirtualMachine{
		ObjectMeta: objectMeta,
		Spec: kubevirtv1.VirtualMachineSpec{
			Running:             running,
			Template:            template,
			DataVolumeTemplates: []kubevirtv1.DataVolumeTemplateSpec{},
		},
	}
	return &VMBuilder{
		vm: vm,
	}
}

func (v *VMBuilder) Name(name string) *VMBuilder {
	v.vm.ObjectMeta.Name = name
	v.vm.ObjectMeta.GenerateName = ""
	v.vm.Spec.Template.ObjectMeta.Labels["harvester.cattle.io/vmName"] = name
	return v
}

func (v *VMBuilder) Namespace(namespace string) *VMBuilder {
	v.vm.ObjectMeta.Namespace = namespace
	return v
}

func (v *VMBuilder) Memory(memory string) *VMBuilder {
	v.vm.Spec.Template.Spec.Domain.Resources.Requests = corev1.ResourceList{
		corev1.ResourceMemory: resource.MustParse(memory),
	}
	return v
}

func (v *VMBuilder) CPU(cores int) *VMBuilder {
	v.vm.Spec.Template.Spec.Domain.CPU.Cores = uint32(cores)
	return v
}

func (v *VMBuilder) Run() *clientv1.VirtualMachine {
	v.vm.Spec.Running = pointer.BoolPtr(true)
	return v.VM()
}

func (v *VMBuilder) VM() *clientv1.VirtualMachine {
	if v.vm.Spec.Template.ObjectMeta.Annotations == nil {
		v.vm.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	sshNames, err := json.Marshal(v.sshNames)
	if err != nil {
		return v.vm
	}
	v.vm.Spec.Template.ObjectMeta.Annotations["harvester.cattle.io/sshNames"] = string(sshNames)
	dataVolumeNames, err := json.Marshal(v.dataVolumeNames)
	if err != nil {
		return v.vm
	}
	v.vm.Spec.Template.ObjectMeta.Annotations["harvester.cattle.io/diskNames"] = string(dataVolumeNames)
	return v.vm
}

func (v *VMBuilder) Update(vm *clientv1.VirtualMachine) *VMBuilder {
	v.vm = vm
	return v
}