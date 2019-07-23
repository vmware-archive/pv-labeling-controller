## pv-labeling-controller

PV labeling controller copies particular set of labels from PVC to associated PV. This is useful for seeing PVs when using `kapp inspect`.

Ideally this functionality would just be done by Kubernetes; however, unlike many other builtin controller labels aren't propagated to PVs from PVCs.

## Building

```
./hack/build.sh
```

To deploy:

```
ytt -f config/ | kbld -f- | kapp deploy -a pv-l-ctrl -f- -c -y
```
