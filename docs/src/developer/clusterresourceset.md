
1. Download or create your manifest.yaml
1. create a configmap from your yaml:
        ```
        kubectl create configmap cm-<manifest>-<semver> --from-file=<manifest>.yaml --dry-run=client -o yaml >  <temporary-name>.yaml
        ```
1. Create a ClusterResourceSet
        ```
                ---
                apiVersion: addons.cluster.x-k8s.io/v1alpha4
                kind: ClusterResourceSet
                metadata:
                name: crs-<manifest>-<semver>
                namespace: default
                spec:
                #mode: "ApplyOnce" #currently default mode
                clusterSelector:
                matchLabels:
                cni: <manifest>-<semver>
                resources:
                - name: cm-<manifest>-<semver>
                kind: ConfigMap
        ```
1. Apply both - examples of ClusterResource Sets are under demo/ClusterResourceSets
