## How to Generate CRD Api Reference Docs
Currently, we are using a doc generator named *gen-crd-api-reference-docs* to generate our API Reference Docs. You can find its project [here](https://github.com/ahmetb/gen-crd-api-reference-docs) on Github.

However, there might be some compatibility issue with *Kubebuilder* which we used for generating CRD template codes. Check [here](https://github.com/ahmetb/gen-crd-api-reference-docs/issues/15) for more information about the compatibility issue.

To generate CRD Api reference docs, some tricks need to be done to bypass the issue mentioned above. This guide will demonstrate what should be done to generate CRD API Reference docs.

### Specify API Package to be generated for

Assume path of the your CRD API Package is `api/v1alpha1`.

Add a `doc.go` under the path `api/v1alpha1` like:
```go
// <Some comments about v1alpha1>
// +groupName=<your-group-name>
package v1alpha1
```

### Specify CRDs which need to be exported
Take type `SampleSet` in `api/v1alpha1/sampleset_types.go` as an example, it's a root crd object.

To export the CRD, add a new comment line with `+genclient`:
```
  //+kubebuilder:object:root=true
  //+kubebuilder:subresource:status
+ //+genclient 

  // SampleSet is the Schema for the SampleSets API
  type SampleSet struct {
```

Some explanation for arguments above:
- config: which config should be used
- template-dir: templates(*.tpl) under the directory will be used to generate docs
- api-dir: where to find all the CRDs (e.g. in our situation, `api/v1alpha1`)
- out-file: path and filename of the generated doc