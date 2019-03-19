## USAGE

add this code to nvidia-device-plugin-daemonset.yaml

```yaml
spec:
    spec:
      containers:
      - env:
        - name: NODENAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: SMTP_LOGIN
          value: YOUR_MAIL_LOGIN
        - name: SMTP_PWD
          value: YOUR_MAIL_PASSWORD
        image: nvidia/k8s-device-plugin:1.10
        imagePullPolicy: IfNotPresent
        livenessProbe:
          exec:
            command:
            - /var/lib/kubelet/device-plugins/dpgpucheck
          failureThreshold: 3
          initialDelaySeconds: 180
          periodSeconds: 5
          successThreshold: 1
          timeoutSeconds: 1
```

BUILD AND COPY `dpgpucheck` to `/var/lib/kubelet/device-plugins/`


# k8s-gpustatuschk
