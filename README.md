# Uffizzi Cluster Operator

Welcome to the Uffizzi Cluster Operator, a Kubernetes operator designed to simplify the creation and management of fully managed virtual clusters. This operator leverages the power of Kubernetes Custom Resource Definitions (CRDs) to extend the Kubernetes API, enabling users to easily provision and configure virtual clusters with custom settings.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
    - [Using Helm](#using-helm)
- [Usage](#usage)
    - [Creating a Uffizzi Cluster](#creating-a-uffizzi-cluster)
      - [Configuring the `UffzziCluster` Custom Resource](#configuring-the-uffzzi-cluster-custom-resource)
- [Contributing](#contributing)
- [License](#license)

## Introduction

The Uffizzi Cluster Operator is designed to facilitate the creation of virtual clusters within a Kubernetes environment. Virtual clusters are isolated Kubernetes clusters that run on top of a physical Kubernetes cluster, providing strong isolation and enabling multi-tenancy capabilities. This operator automates the provisioning of virtual clusters, their associated resources, and configurations, making it easier for administrators and developers to manage complex Kubernetes environments.

## Features

- **Uffizzi Cluster Creation**: Automate the creation of virtual clusters with custom configurations.
- **Ingress Configuration**: Expose Ingress for the virtual cluster. Ingresses created inside a UffizziCluster are automatically exposed through the host cluster. 
- **Resource Management**: Configure resource quotas, limit ranges, and other Kubernetes resources for the virtual cluster.
- **Sleep Mode**: Suspend the virtual cluster if it is not being used for a certain period, optimizing resource usage.

## Prerequisites

Before installing the Uffizzi Cluster Operator, ensure you have the following:

- A Kubernetes cluster (version 1.16 or later) with Helm 3 installed.
- Access to the cluster with administrator privileges.

## Installation

### Using Helm

The Uffizzi Cluster Operator can be installed using Helm, a package manager for Kubernetes. This method simplifies the deployment and management of the operator within your cluster.

```bash
helm dep update ./chart
helm install ./chart
```

## Usage

### Creating a Uffizzi Cluster

To create a uffizzi virtual cluster, you need to define an `UffizziCluster` custom resource in your Kubernetes cluster. Here is an example manifest:

```yaml
kind: UffizziCluster
apiVersion: uffizzi.com/v1alpha1
metadata:
  name: test-ucluster
spec:
  manifests: |
    apiVersion: v1
    kind: Pod
    metadata:
      name: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80

```

Apply this manifest using `kubectl`:

```bash
kubectl apply -f test-ucluster.yaml
```

#### Configuring the `UffzziCluster` Custom Resource

```yaml
apiVersion: uffizzi.cloud/v1alpha1
kind: UffizziCluster
metadata:
  # Metadata for the UffizziCluster resource, including name, namespace, labels, and annotations.
spec:
  distro: 
    # The Kubernetes distribution used for the virtual cluster. Default is "k3s".
    # Supported values: "k3s", "k8s".
  nodeSelectorTemplate: 
    # Template for node selection constraints.
  nodeSelector: 
    # Node labels for pod assignment.
  tolerations: 
    # Tolerations for pod assignment.
  apiServer: 
    # Configuration for the API server of the virtual cluster.
    image: 
      # The container image for the API server.
  ingress: 
    # Ingress configuration for the virtual cluster.
    host: 
      # The hostname for the ingress.
    class: 
      # The ingress class to use.
  helm: 
    # List of Helm charts to be installed in the virtual cluster.
    - chart: 
        name: 
          # Name of the Helm chart.
        repo: 
          # Repository URL of the Helm chart.
        Version: 
          # Version of the Helm chart. Optional.
      values: 
        # Custom values for the Helm chart installation. Optional.
      release: 
        name: 
          # Name of the Helm release.
        namespace: 
          # Namespace for the Helm release.
  manifests: 
    # Raw Kubernetes manifests to apply to the virtual cluster. Optional.
  resourceQuota: 
    # Resource quota configuration for the virtual cluster.
    enabled: 
      # Whether resource quotas are enabled. Default is true.
    requests: 
      # Minimum resource requests.
      cpu: 
        # CPU request. Default is "0.5".
      memory: 
        # Memory request. Default is "1Gi".
      ephemeralStorage: 
        # Ephemeral storage request. Default is "5Gi".
      storage: 
        # Storage request. Default is "10Gi".
    limits: 
      # Maximum resource limits.
      cpu: 
        # CPU limit. Default is "0.5".
      memory: 
        # Memory limit. Default is "8Gi".
      ephemeralStorage: 
        # Ephemeral storage limit. Default is "5Gi".
    services: 
      # Service quotas.
      nodePorts: 
        # Number of NodePort services. Default is 0.
      loadBalancers: 
        # Number of LoadBalancer services. Default is 3.
    count: 
      # Count quotas for various resources.
      pods: 
        # Number of pods. Default is 20.
      services: 
        # Number of services. Default is 10.
      configMaps: 
        # Number of ConfigMaps. Default is 20.
      secrets: 
        # Number of secrets. Default is 20.
      persistentVolumeClaims: 
        # Number of PVCs. Default is 10.
      endpoints: 
        # Number of endpoints. Default is 10.
  limitRange: 
    # Limit range configuration for the virtual cluster.
    enabled: 
      # Whether limit ranges are enabled. Default is true.
    default: 
      # Default limits for resources.
      cpu: 
        # Default CPU limit. Default is "0.5".
      memory: 
        # Default memory limit. Default is "1Gi".
      ephemeralStorage: 
        # Default ephemeral storage limit. Default is "8Gi".
    defaultRequest: 
      # Default requests for resources.
      cpu: 
        # Default CPU request. Default is "0.1".
      memory: 
        # Default memory request. Default is "128Mi".
      ephemeralStorage: 
        # Default ephemeral storage request. Default is "1Gi".
  sleep: 
    # Whether the virtual cluster is in sleep mode. Default is false.
  storage: 
    # Storage configuration for the virtual cluster.
    persistence: 
      # Whether persistence is enabled. Default is true.
    size: 
      # Size of the storage. Default is "5Gi".
  externalDatastore: 
    # Type of external datastore used. Default is "sqlite".
    # Supported values: "etcd", "sqlite".
status:
  # Observed state of the UffizziCluster.
```

## Contributing

We welcome contributions from the community! If you're interested in contributing to the Uffizzi Cluster Operator, please check out our [contributing guidelines](CONTRIBUTING.md).

## License

The Uffizzi Cluster Operator is open-source software licensed under the [Apache 2.0 License](LICENSE).
