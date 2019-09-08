package strace

import (
	"fmt"
	"io"
	"io/ioutil"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	batchv1typed "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1typed "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	TracePrefix = "ktrace-"

	TraceIDLabelKey = "sqr/ktrace-id"
	TraceLabelKey   = "sqr/ktrace"
)

var (
	KStrace_ImageNameTag = "registry.cn-hangzhou.aliyuncs.com/test_dev/sqr:v1.2"
)

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }

type StraceJobClient struct {
	JobClient    batchv1typed.JobInterface
	ConfigClient corev1typed.ConfigMapInterface
	outStream    io.Writer
}

func (t *StraceJobClient) WithOutStream(o io.Writer) {
	if o == nil {
		t.outStream = ioutil.Discard
	}
	t.outStream = o
}

type StraceJobStatus string

const (
	TraceJobRunning   StraceJobStatus = "Running"
	TraceJobCompleted StraceJobStatus = "Completed"
	TraceJobFailed    StraceJobStatus = "Failed"
	TraceJobUnknown   StraceJobStatus = "Unknown"
)

// StraceJob is a container of info needed to create the job responsible for tracing.
type StraceJob struct {
	Name          string
	ID            types.UID
	Namespace     string
	ContainerName string
	PodName       string
	PodUID        string
	Hostname      string

	Program string

	StartTime metav1.Time
	Status    StraceJobStatus
}

func (t *StraceJobClient) CreateJob(nj StraceJob) (*batchv1.Job, error) {

	bpfTraceCmd := []string{
		"/root/kstrace_exec_linux",
		"-f=/kstrace/root.bt",
		"-c=" + nj.ContainerName,
		"-p=" + nj.PodUID,
	}

	commonMeta := metav1.ObjectMeta{
		Name:      nj.Name,
		Namespace: nj.Namespace,
		Labels: map[string]string{
			TraceLabelKey:   nj.Name,
			TraceIDLabelKey: string(nj.ID),
		},
		Annotations: map[string]string{
			TraceLabelKey:   nj.Name,
			TraceIDLabelKey: string(nj.ID),
		},
	}

	cm := &apiv1.ConfigMap{
		ObjectMeta: commonMeta,
		Data: map[string]string{
			"root.bt": nj.Program,
		},
	}

	job := &batchv1.Job{
		ObjectMeta: commonMeta,
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: commonMeta,
				Spec: apiv1.PodSpec{
					HostPID: true,
					Volumes: []apiv1.Volume{
						apiv1.Volume{
							Name: "kstrace",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: cm.Name,
									},
								},
							},
						},
						apiv1.Volume{
							Name: "usr-src-host",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/usr/src",
								},
							},
						},
						apiv1.Volume{
							Name: "modules-host",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/lib/modules",
								},
							},
						},
						apiv1.Volume{
							Name: "sys-host",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/sys",
								},
							},
						},
					},

					Containers: []apiv1.Container{
						apiv1.Container{
							Name:    nj.Name,
							Image:   KStrace_ImageNameTag,
							Command: bpfTraceCmd,
							TTY:     true,
							Stdin:   true,
							VolumeMounts: []apiv1.VolumeMount{
								apiv1.VolumeMount{
									Name:      "kstrace",
									MountPath: "/kstrace",
									ReadOnly:  true,
								},
								apiv1.VolumeMount{
									Name:      "sys-host",
									MountPath: "/sys",
									ReadOnly:  true,
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: boolPtr(true),
							},
						},
					},
					RestartPolicy: "Never",
					Affinity: &apiv1.Affinity{
						NodeAffinity: &apiv1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &apiv1.NodeSelector{
								NodeSelectorTerms: []apiv1.NodeSelectorTerm{
									apiv1.NodeSelectorTerm{
										MatchExpressions: []apiv1.NodeSelectorRequirement{
											apiv1.NodeSelectorRequirement{
												Key:      "kubernetes.io/hostname",
												Operator: apiv1.NodeSelectorOpIn,
												Values:   []string{nj.Hostname},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts,
		apiv1.VolumeMount{
			Name:      "usr-src-host",
			MountPath: "/usr/src",
			ReadOnly:  true,
		},
		apiv1.VolumeMount{
			Name:      "modules-host",
			MountPath: "/lib/modules",
			ReadOnly:  true,
		})
	if _, err := t.ConfigClient.Create(cm); err != nil {
		return nil, err
	}
	return t.JobClient.Create(job)
}

func (t *StraceJobClient) DeleteJob(name string) error {
	dp := metav1.DeletePropagationForeground
	err := t.JobClient.Delete(name, &metav1.DeleteOptions{
		GracePeriodSeconds: int64Ptr(0),
		PropagationPolicy:  &dp,
	})

	if err != nil {
		return err
	}
	fmt.Fprintf(t.outStream, "trace job %s deleted\n", name)

	err = t.ConfigClient.Delete(name, nil)
	if err != nil {
		return err
	}
	fmt.Fprintf(t.outStream, "trace configuration %s deleted\n", name)
	return nil
}
