package controllers

import (
	"context"
	"encoding/json"
	uclusteruffizzicomv1alpha1 "github.com/UffizziCloud/uffizzi-cluster-operator/api/v1alpha1"
	fluxhelmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	fluxsourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/pkg/errors"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *UffizziClusterReconciler) createLoftHelmRepo(ctx context.Context, req ctrl.Request) error {
	return r.createHelmRepo(ctx, LOFT_HELM_REPO, req.Namespace, LOFT_CHART_REPO_URL)
}

func (r *UffizziClusterReconciler) deleteLoftHelmRepo(ctx context.Context, req ctrl.Request) error {
	return r.deleteHelmRepo(ctx, LOFT_HELM_REPO, req.Namespace)
}

func (r *UffizziClusterReconciler) upsertVClusterK3sHelmRelease(update bool, ctx context.Context, uCluster *uclusteruffizzicomv1alpha1.UffizziCluster) (*fluxhelmv2beta1.HelmRelease, error) {
	helmReleaseName := BuildVClusterHelmReleaseName(uCluster)
	var (
		VClusterIngressHostname     = BuildVClusterIngressHost(uCluster)
		OutKubeConfigServerArgValue = ""
	)

	if VClusterIngressHostname != "" {
		OutKubeConfigServerArgValue = "https://" + VClusterIngressHostname
	}

	vclusterK3sHelmValues := VClusterK3S{
		VCluster: VClusterContainer{
			Image: "rancher/k3s:v1.27.3-k3s1",
		},
		Init:    VClusterInit{},
		FsGroup: 12345,
		Ingress: VClusterIngress{
			Enabled: true,
			Host:    VClusterIngressHostname,
			Annotations: map[string]string{
				"app.uffizzi.com/ingress-sync": "true",
			},
		},
		Isolation: VClusterIsolation{
			Enabled:             true,
			PodSecurityStandard: "baseline",
			ResourceQuota: VClusterResourceQuota{
				Enabled: true,
				Quota: VClusterResourceQuotaDefiniton{
					RequestsCpu:                 "2.5",
					RequestsMemory:              "10Gi",
					RequestsEphemeralStorage:    "15Gi",
					RequestsStorage:             "10Gi",
					LimitsCpu:                   "10",
					LimitsMemory:                "15Gi",
					LimitsEphemeralStorage:      "30Gi",
					ServicesLoadbalancers:       5,
					ServicesNodePorts:           0,
					CountEndpoints:              40,
					CountConfigmaps:             100,
					CountPersistentVolumeClaims: 40,
					CountPods:                   40,
					CountSecrets:                100,
					CountServices:               40,
				},
			},
			LimitRange: VClusterLimitRange{
				Enabled: true,
				Default: LimitRangeResources{
					Cpu:              "1",
					Memory:           "512Mi",
					EphemeralStorage: "8Gi",
				},
				DefaultRequest: LimitRangeResources{
					Cpu:              "100m",
					Memory:           "128Mi",
					EphemeralStorage: "3Gi",
				},
			},
			NetworkPolicy: VClusterNetworkPolicy{
				Enabled: true,
			},
		},
		NodeSelector: VClusterNodeSelector{
			SandboxGKEIORuntime: "gvisor",
		},
		SecurityContext: VClusterSecurityContext{
			Capabilities: VClusterSecurityContextCapabilities{
				Drop: []string{"all"},
			},
		},
		Tolerations: []VClusterToleration{
			{
				Key:      "sandbox.gke.io/runtime",
				Effect:   "NoSchedule",
				Operator: "Exists",
			},
		},
		Plugin: VClusterPlugins{
			VClusterPlugin{
				Image:           "uffizzi/ucluster-sync-plugin:v0.2.4",
				ImagePullPolicy: "IfNotPresent",
				Rbac: VClusterRbac{
					Role: VClusterRbacRole{
						ExtraRules: []VClusterRbacRule{
							{
								ApiGroups: []string{"networking.k8s.io"},
								Resources: []string{"ingresses"},
								Verbs:     []string{"create", "delete", "patch", "update", "get", "list", "watch"},
							},
						},
					},
					ClusterRole: VClusterRbacClusterRole{
						ExtraRules: []VClusterRbacRule{
							{
								ApiGroups: []string{"apiextensions.k8s.io"},
								Resources: []string{"customresourcedefinitions"},
								Verbs:     []string{"patch", "update", "get", "list", "watch"},
							},
						},
					},
				},
			},
		},
		Syncer: VClusterSyncer{
			KubeConfigContextName: helmReleaseName,
			ExtraArgs: []string{
				"--enforce-toleration=sandbox.gke.io/runtime:NoSchedule",
				"--node-selector=sandbox.gke.io/runtime=gvisor",
				"--enforce-node-selector",
			},
		},
		Sync: VClusterSync{
			Ingresses: EnabledBool{
				Enabled: false,
			},
		},
	}

	if uCluster.Spec.APIServer.Image != "" {
		vclusterK3sHelmValues.VCluster.Image = uCluster.Spec.APIServer.Image
	}

	if uCluster.Spec.Ingress.Host != "" {
		vclusterK3sHelmValues.Plugin.UffizziClusterSyncPlugin.Env = []VClusterContainerEnv{
			{
				Name:  "VCLUSTER_INGRESS_HOST",
				Value: VClusterIngressHostname,
			},
		}
	}

	if uCluster.Spec.ResourceQuota != nil {
		// map uCluster.Spec.ResourceQuota to vclusterK3sHelmValues.Isolation.ResourceQuota
		q := *uCluster.Spec.ResourceQuota
		qHelmValues := vclusterK3sHelmValues.Isolation.ResourceQuota
		// enabled
		qHelmValues.Enabled = q.Enabled
		//requests
		qHelmValues.Quota.RequestsMemory = q.Requests.Memory
		qHelmValues.Quota.RequestsCpu = q.Requests.CPU
		qHelmValues.Quota.RequestsEphemeralStorage = q.Requests.EphemeralStorage
		qHelmValues.Quota.RequestsStorage = q.Requests.Storage
		// limits
		qHelmValues.Quota.LimitsMemory = q.Limits.Memory
		qHelmValues.Quota.LimitsCpu = q.Limits.CPU
		qHelmValues.Quota.LimitsEphemeralStorage = q.Limits.EphemeralStorage
		// services
		qHelmValues.Quota.ServicesNodePorts = q.Services.NodePorts
		qHelmValues.Quota.ServicesLoadbalancers = q.Services.LoadBalancers
		// count
		qHelmValues.Quota.CountPods = q.Count.Pods
		qHelmValues.Quota.CountServices = q.Count.Services
		qHelmValues.Quota.CountPersistentVolumeClaims = q.Count.PersistentVolumeClaims
		qHelmValues.Quota.CountConfigmaps = q.Count.ConfigMaps
		qHelmValues.Quota.CountSecrets = q.Count.Secrets
		qHelmValues.Quota.CountEndpoints = q.Count.Endpoints
		// set it back
		vclusterK3sHelmValues.Isolation.ResourceQuota = qHelmValues
	}

	if uCluster.Spec.LimitRange != nil {
		// same for limit range
		lr := uCluster.Spec.LimitRange
		lrHelmValues := vclusterK3sHelmValues.Isolation.LimitRange
		// enabled
		lrHelmValues.Enabled = lr.Enabled
		// default
		lrHelmValues.Default.Cpu = lr.Default.CPU
		lrHelmValues.Default.Memory = lr.Default.Memory
		lrHelmValues.Default.EphemeralStorage = lr.Default.EphemeralStorage
		// default requests
		lrHelmValues.DefaultRequest.Cpu = lr.DefaultRequest.CPU
		lrHelmValues.DefaultRequest.Memory = lr.DefaultRequest.Memory
		lrHelmValues.DefaultRequest.EphemeralStorage = lr.DefaultRequest.EphemeralStorage
		// set it back
		vclusterK3sHelmValues.Isolation.LimitRange = lrHelmValues
	}

	vclusterK3sHelmValues.Syncer.ExtraArgs = append(vclusterK3sHelmValues.Syncer.ExtraArgs,
		"--tls-san="+VClusterIngressHostname,
		"--out-kube-config-server="+OutKubeConfigServerArgValue,
	)

	if len(uCluster.Spec.Helm) > 0 {
		vclusterK3sHelmValues.Init.Helm = uCluster.Spec.Helm
	}

	if uCluster.Spec.Manifests != nil {
		vclusterK3sHelmValues.Init.Manifests = *uCluster.Spec.Manifests
	}

	// marshal HelmValues struct to JSON
	helmValuesRaw, err := json.Marshal(vclusterK3sHelmValues)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal HelmValues struct to JSON")
	}

	// Create the apiextensionsv1.JSON instance with the raw data
	helmValuesJSONObj := v1.JSON{Raw: helmValuesRaw}

	// Create a new HelmRelease
	newHelmRelease := &fluxhelmv2beta1.HelmRelease{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      helmReleaseName,
			Namespace: uCluster.Namespace,
		},
		Spec: fluxhelmv2beta1.HelmReleaseSpec{
			Upgrade: &fluxhelmv2beta1.Upgrade{
				Force: false,
			},
			Chart: fluxhelmv2beta1.HelmChartTemplate{
				Spec: fluxhelmv2beta1.HelmChartTemplateSpec{
					Chart:   VCLUSTER_CHART_K3S,
					Version: VCLUSTER_CHART_K3S_VERSION,
					SourceRef: fluxhelmv2beta1.CrossNamespaceObjectReference{
						Kind:      "HelmRepository",
						Name:      LOFT_HELM_REPO,
						Namespace: uCluster.Namespace,
					},
				},
			},
			ReleaseName: helmReleaseName,
			Values:      &helmValuesJSONObj,
		},
	}

	if err := controllerutil.SetControllerReference(uCluster, newHelmRelease, r.Scheme); err != nil {
		return nil, errors.Wrap(err, "failed to set controller reference")
	}
	// get the helm release spec in string
	newHelmReleaseSpecBytes, err := json.Marshal(newHelmRelease.Spec)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal current spec")
	}
	newHelmReleaseSpec := string(newHelmReleaseSpecBytes)
	// upsert
	if !update && uCluster.Status.LastAppliedHelmReleaseSpec == nil {
		if err := r.Create(ctx, newHelmRelease); err != nil {
			return nil, errors.Wrap(err, "failed to create HelmRelease")
		}
		patch := client.MergeFrom(uCluster.DeepCopy())
		uCluster.Status.LastAppliedHelmReleaseSpec = &newHelmReleaseSpec
		if err := r.Status().Patch(ctx, uCluster, patch); err != nil {
			return nil, errors.Wrap(err, "Failed to update the default UffizziCluster lastAppliedHelmReleaseSpec")
		}

	} else if uCluster.Status.LastAppliedHelmReleaseSpec != nil {
		// create helm release if there is no existing helm release to update
		if update && *uCluster.Status.LastAppliedHelmReleaseSpec != newHelmReleaseSpec {
			if err := r.updateHelmRelease(newHelmRelease, uCluster, ctx); err != nil {
				return nil, errors.Wrap(err, "failed to update HelmRelease")
			}
			return nil, errors.Wrap(err, "couldn't update HelmRelease as LastAppliedHelmReleaseSpec does not exist on resource")
		}
	}

	return newHelmRelease, nil
}

func (r *UffizziClusterReconciler) upsertVClusterK8sHelmRelease(update bool, ctx context.Context, uCluster *uclusteruffizzicomv1alpha1.UffizziCluster) (*fluxhelmv2beta1.HelmRelease, error) {
	helmReleaseName := BuildVClusterHelmReleaseName(uCluster)
	var (
		VClusterIngressHostname     = BuildVClusterIngressHost(uCluster)
		OutKubeConfigServerArgValue = "https://" + VClusterIngressHostname
	)

	vclusterHelmValues := VClusterK8S{
		APIServer: VClusterK8SAPIServer{
			Image: "registry.k8s.io/kube-apiserver:v1.26.1",
			Resources: VClusterContainerResources{
				Requests: VClusterContainerResourcesRequests{
					Cpu:    "40m",
					Memory: "300Mi",
				},
			},
		},
		Init:    VClusterInit{},
		FsGroup: 12345,
		Ingress: VClusterIngress{
			Enabled: true,
			Host:    VClusterIngressHostname,
			Annotations: map[string]string{
				"app.uffizzi.com/ingress-sync": "true",
			},
		},
		Isolation: VClusterIsolation{
			Enabled:             true,
			PodSecurityStandard: "baseline",
			ResourceQuota: VClusterResourceQuota{
				Enabled: true,
				Quota: VClusterResourceQuotaDefiniton{
					RequestsCpu:                 "10",
					RequestsMemory:              "20Gi",
					RequestsEphemeralStorage:    "60Gi",
					RequestsStorage:             "100Gi",
					LimitsCpu:                   "20",
					LimitsMemory:                "40Gi",
					LimitsEphemeralStorage:      "160Gi",
					ServicesLoadbalancers:       5,
					ServicesNodePorts:           0,
					CountEndpoints:              40,
					CountConfigmaps:             100,
					CountPersistentVolumeClaims: 20,
					CountPods:                   20,
					CountSecrets:                100,
					CountServices:               20,
				},
			},
			LimitRange: VClusterLimitRange{
				Enabled: true,
				Default: LimitRangeResources{
					Cpu:              "1",
					Memory:           "512Mi",
					EphemeralStorage: "8Gi",
				},
				DefaultRequest: LimitRangeResources{
					Cpu:              "100m",
					Memory:           "128Mi",
					EphemeralStorage: "3Gi",
				},
			},
		},
		NodeSelector: VClusterNodeSelector{
			SandboxGKEIORuntime: "gvisor",
		},
		SecurityContext: VClusterSecurityContext{
			Capabilities: VClusterSecurityContextCapabilities{
				Drop: []string{"all"},
			},
		},
		Tolerations: []VClusterToleration{
			{
				Key:      "sandbox.gke.io/runtime",
				Effect:   "NoSchedule",
				Operator: "Exists",
			},
		},
		Plugin: VClusterPlugins{
			VClusterPlugin{
				Image:           "uffizzi/ucluster-sync-plugin:v0.2.4",
				ImagePullPolicy: "IfNotPresent",
				Rbac: VClusterRbac{
					Role: VClusterRbacRole{
						ExtraRules: []VClusterRbacRule{
							{
								ApiGroups: []string{"networking.k8s.io"},
								Resources: []string{"ingresses"},
								Verbs:     []string{"create", "delete", "patch", "update", "get", "list", "watch"},
							},
						},
					},
					ClusterRole: VClusterRbacClusterRole{
						ExtraRules: []VClusterRbacRule{
							{
								ApiGroups: []string{"apiextensions.k8s.io"},
								Resources: []string{"customresourcedefinitions"},
								Verbs:     []string{"patch", "update", "get", "list", "watch"},
							},
						},
					},
				},
			},
		},
		Syncer: VClusterSyncer{
			KubeConfigContextName: helmReleaseName,
			ExtraArgs: []string{
				"--enforce-toleration=sandbox.gke.io/runtime:NoSchedule",
				"--node-selector=sandbox.gke.io/runtime=gvisor",
				"--enforce-node-selector",
			},
		},
		Sync: VClusterSync{
			Ingresses: EnabledBool{
				Enabled: false,
			},
		},
	}

	if uCluster.Spec.APIServer.Image != "" {
		vclusterHelmValues.APIServer.Image = uCluster.Spec.APIServer.Image
	}

	if uCluster.Spec.Ingress.Host != "" {
		vclusterHelmValues.Plugin.UffizziClusterSyncPlugin.Env = []VClusterContainerEnv{
			{
				Name:  "VCLUSTER_INGRESS_HOST",
				Value: VClusterIngressHostname,
			},
		}
	}

	if uCluster.Spec.ResourceQuota != nil {
		// map uCluster.Spec.ResourceQuota to vclusterHelmValues.Isolation.ResourceQuota
		q := *uCluster.Spec.ResourceQuota
		qHelmValues := vclusterHelmValues.Isolation.ResourceQuota
		// enabled
		qHelmValues.Enabled = q.Enabled
		//requests
		qHelmValues.Quota.RequestsMemory = q.Requests.Memory
		qHelmValues.Quota.RequestsCpu = q.Requests.CPU
		qHelmValues.Quota.RequestsEphemeralStorage = q.Requests.EphemeralStorage
		qHelmValues.Quota.RequestsStorage = q.Requests.Storage
		// limits
		qHelmValues.Quota.LimitsMemory = q.Limits.Memory
		qHelmValues.Quota.LimitsCpu = q.Limits.CPU
		qHelmValues.Quota.LimitsEphemeralStorage = q.Limits.EphemeralStorage
		// services
		qHelmValues.Quota.ServicesNodePorts = q.Services.NodePorts
		qHelmValues.Quota.ServicesLoadbalancers = q.Services.LoadBalancers
		// count
		qHelmValues.Quota.CountPods = q.Count.Pods
		qHelmValues.Quota.CountServices = q.Count.Services
		qHelmValues.Quota.CountPersistentVolumeClaims = q.Count.PersistentVolumeClaims
		qHelmValues.Quota.CountConfigmaps = q.Count.ConfigMaps
		qHelmValues.Quota.CountSecrets = q.Count.Secrets
		qHelmValues.Quota.CountEndpoints = q.Count.Endpoints
		// set it back
		vclusterHelmValues.Isolation.ResourceQuota = qHelmValues
	}

	if uCluster.Spec.LimitRange != nil {
		// same for limit range
		lr := uCluster.Spec.LimitRange
		lrHelmValues := vclusterHelmValues.Isolation.LimitRange
		// enabled
		lrHelmValues.Enabled = lr.Enabled
		// default
		lrHelmValues.Default.Cpu = lr.Default.CPU
		lrHelmValues.Default.Memory = lr.Default.Memory
		lrHelmValues.Default.EphemeralStorage = lr.Default.EphemeralStorage
		// default requests
		lrHelmValues.DefaultRequest.Cpu = lr.DefaultRequest.CPU
		lrHelmValues.DefaultRequest.Memory = lr.DefaultRequest.Memory
		lrHelmValues.DefaultRequest.EphemeralStorage = lr.DefaultRequest.EphemeralStorage
		// set it back
		vclusterHelmValues.Isolation.LimitRange = lrHelmValues
	}

	vclusterHelmValues.Syncer.ExtraArgs = append(vclusterHelmValues.Syncer.ExtraArgs,
		"--tls-san="+VClusterIngressHostname,
		"--out-kube-config-server="+OutKubeConfigServerArgValue,
	)

	if len(uCluster.Spec.Helm) > 0 {
		vclusterHelmValues.Init.Helm = uCluster.Spec.Helm
	}

	if uCluster.Spec.Manifests != nil {
		vclusterHelmValues.Init.Manifests = *uCluster.Spec.Manifests
	}

	// marshal HelmValues struct to JSON
	helmValuesRaw, err := json.Marshal(vclusterHelmValues)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal HelmValues struct to JSON")
	}

	// Create the apiextensionsv1.JSON instance with the raw data
	helmValuesJSONObj := v1.JSON{Raw: helmValuesRaw}

	// Create a new HelmRelease
	newHelmRelease := &fluxhelmv2beta1.HelmRelease{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      helmReleaseName,
			Namespace: uCluster.Namespace,
		},
		Spec: fluxhelmv2beta1.HelmReleaseSpec{
			Upgrade: &fluxhelmv2beta1.Upgrade{
				Force: false,
			},
			Chart: fluxhelmv2beta1.HelmChartTemplate{
				Spec: fluxhelmv2beta1.HelmChartTemplateSpec{
					Chart:   VCLUSTER_CHART_K8S,
					Version: VCLUSTER_CHART_K8S_VERSION,
					SourceRef: fluxhelmv2beta1.CrossNamespaceObjectReference{
						Kind:      "HelmRepository",
						Name:      LOFT_HELM_REPO,
						Namespace: uCluster.Namespace,
					},
				},
			},
			ReleaseName: helmReleaseName,
			Values:      &helmValuesJSONObj,
		},
	}

	if err := controllerutil.SetControllerReference(uCluster, newHelmRelease, r.Scheme); err != nil {
		return nil, errors.Wrap(err, "failed to set controller reference")
	}
	// get the helm release spec in string
	newHelmReleaseSpecBytes, err := json.Marshal(newHelmRelease.Spec)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal current spec")
	}
	newHelmReleaseSpec := string(newHelmReleaseSpecBytes)
	// upsert
	if !update && uCluster.Status.LastAppliedHelmReleaseSpec == nil {
		if err := r.Create(ctx, newHelmRelease); err != nil {
			return nil, errors.Wrap(err, "failed to create HelmRelease")
		}
		patch := client.MergeFrom(uCluster.DeepCopy())
		uCluster.Status.LastAppliedHelmReleaseSpec = &newHelmReleaseSpec
		if err := r.Status().Patch(ctx, uCluster, patch); err != nil {
			return nil, errors.Wrap(err, "Failed to update the default UffizziCluster lastAppliedHelmReleaseSpec")
		}

	} else if uCluster.Status.LastAppliedHelmReleaseSpec != nil {
		// create helm release if there is no existing helm release to update
		if update && *uCluster.Status.LastAppliedHelmReleaseSpec != newHelmReleaseSpec {
			if err := r.updateHelmRelease(newHelmRelease, uCluster, ctx); err != nil {
				return nil, errors.Wrap(err, "failed to update HelmRelease")
			}
			return nil, errors.Wrap(err, "couldn't update HelmRelease as LastAppliedHelmReleaseSpec does not exist on resource")
		}
	}

	return newHelmRelease, nil
}

func (r *UffizziClusterReconciler) updateHelmRelease(newHelmRelease *fluxhelmv2beta1.HelmRelease, uCluster *uclusteruffizzicomv1alpha1.UffizziCluster, ctx context.Context) error {
	existingHelmRelease := &fluxhelmv2beta1.HelmRelease{}
	existingHelmReleaseNN := types.NamespacedName{
		Name:      newHelmRelease.Name,
		Namespace: newHelmRelease.Namespace,
	}
	if err := r.Get(ctx, existingHelmReleaseNN, existingHelmRelease); err != nil {
		return errors.Wrap(err, "failed to find HelmRelease")
	}
	// check if the helm release is already progressing, if so, do not update
	if existingHelmRelease.Status.Conditions != nil {
		for _, condition := range existingHelmRelease.Status.Conditions {
			if condition.Type == fluxhelmv2beta1.ReleasedCondition && condition.Status == "Unknown" && condition.Reason == "Progressing" {
				return nil
			}
		}
	}

	newHelmRelease.Spec.Upgrade = &fluxhelmv2beta1.Upgrade{
		Force: true,
	}
	existingHelmRelease.Spec = newHelmRelease.Spec
	if err := r.Update(ctx, existingHelmRelease); err != nil {
		return errors.Wrap(err, "error while updating helm release")
	}
	// update the lastAppliedConfig
	updatedSpecBytes, err := json.Marshal(uCluster.Spec)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal current spec")
	}
	updatedHelmReleaseSpecBytes, err := json.Marshal(existingHelmRelease.Spec)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal current spec")
	}
	updatedSpec := string(updatedSpecBytes)
	updatedHelmReleaseSpec := string(updatedHelmReleaseSpecBytes)
	patch := client.MergeFrom(uCluster.DeepCopy())
	uCluster.Status.LastAppliedConfiguration = &updatedSpec
	uCluster.Status.LastAppliedHelmReleaseSpec = &updatedHelmReleaseSpec
	if err := r.Status().Patch(ctx, uCluster, patch); err != nil {
		return errors.Wrap(err, "Failed to update the default UffizziCluster lastAppliedConfig")
	}
	return nil
}

func (r *UffizziClusterReconciler) createHelmRepo(ctx context.Context, name, namespace, url string) error {
	// Create HelmRepository in the same namespace as the HelmRelease
	helmRepo := &fluxsourcev1.HelmRepository{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: fluxsourcev1.HelmRepositorySpec{
			URL: url,
		},
	}

	err := r.Create(ctx, helmRepo)
	return err
}

func (r *UffizziClusterReconciler) deleteHelmRepo(ctx context.Context, name, namespace string) error {
	// Create HelmRepository in the same namespace as the HelmRelease
	helmRepo := &fluxsourcev1.HelmRepository{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := r.Delete(ctx, helmRepo)
	return err
}
