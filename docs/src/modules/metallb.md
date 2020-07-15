# MetalLB
[A MetalLB controller](https://github.com/simonswine/hcloud-metallb-floater) is automatically created.
The controller assigns existing floating IPs to nodes.

## Setting up a floating IP 
To provide MetalLB with a floating IP, you have to go through a few steps.
First you create a floating IP in the console, then you inform Metallb that it can use the ip and finally you create the service.

**Create floating IP**

1. Go to the cloud console
2. Click "Add Floating IP" in the floating IPs section
3. Select the cluster location and add

**Inform MetalLB**

1. `kubectl edit configmaps -n kube-system metallb-config`
2. Update the data section to:
```yaml
  config: |-
    "address-pools":
    - "addresses":
      - "<apiserver-ip>/32"
      "name": "kube-apiserver"
      "protocol": "layer2"
    - "addresses":
      - "<service-ip>/32"
      "name": "test-service-ip"
      "protocol": "layer2"
```

**Create Service**

When a LoadBalancer service requests an IP, MetalLB will automatically assign the IP to that service.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: LoadBalancer
  selector:
    app: nginx
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
```

## IPv6
TODO