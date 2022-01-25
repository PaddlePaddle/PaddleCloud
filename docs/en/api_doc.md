<p>Packages:</p>
<ul>
<li>
<a href="#batch.paddlepaddle.org%2fv1">batch.paddlepaddle.org/v1</a>
</li>
<li>
<a href="#batch.paddlepaddle.org%2fv1alpha1">batch.paddlepaddle.org/v1alpha1</a>
</li>
</ul>
<h2 id="batch.paddlepaddle.org/v1">batch.paddlepaddle.org/v1</h2>
<p>
<p>Package v1 contains PaddleJob</p>
</p>
Resource Types:
<ul><li>
<a href="#batch.paddlepaddle.org/v1.PaddleJob">PaddleJob</a>
</li></ul>
<h3 id="batch.paddlepaddle.org/v1.PaddleJob">PaddleJob
</h3>
<p>
<p>PaddleJob is the Schema for the paddlejobs API</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
batch.paddlepaddle.org/v1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>PaddleJob</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobSpec">
PaddleJobSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>cleanPodPolicy</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.CleanPodPolicy">
CleanPodPolicy
</a>
</em>
</td>
<td>
<p>CleanPodPolicy defines whether to clean pod after job finished</p>
</td>
</tr>
<tr>
<td>
<code>schedulingPolicy</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.SchedulingPolicy">
SchedulingPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchedulingPolicy defines the policy related to scheduling, for volcano</p>
</td>
</tr>
<tr>
<td>
<code>intranet</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.Intranet">
Intranet
</a>
</em>
</td>
<td>
<p>Intranet defines the communication mode inter pods : PodIP, Service or Host</p>
</td>
</tr>
<tr>
<td>
<code>withGloo</code></br>
<em>
int
</em>
</td>
<td>
<p>WithGloo indicate whether enable gloo, 0/1/2 for disable/enable for worker/enable for server</p>
</td>
</tr>
<tr>
<td>
<code>sampleSetRef</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.SampleSetRef">
SampleSetRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SampleSetRef defines the sample data set used for training and its mount path in worker pods</p>
</td>
</tr>
<tr>
<td>
<code>ps</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.ResourceSpec">
ResourceSpec
</a>
</em>
</td>
<td>
<p>PS[erver] describes the spec of server base on pod template</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.ResourceSpec">
ResourceSpec
</a>
</em>
</td>
<td>
<p>Worker describes the spec of worker base on pod template</p>
</td>
</tr>
<tr>
<td>
<code>heter</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.ResourceSpec">
ResourceSpec
</a>
</em>
</td>
<td>
<p>Heter describes the spec of heter worker base on pod temlate</p>
</td>
</tr>
<tr>
<td>
<code>elastic</code></br>
<em>
int
</em>
</td>
<td>
<p>Elastic indicate the elastic level</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobStatus">
PaddleJobStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1.CleanPodPolicy">CleanPodPolicy
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobSpec">PaddleJobSpec</a>)
</p>
<p>
</p>
<h3 id="batch.paddlepaddle.org/v1.ElasticStatus">ElasticStatus
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobStatus">PaddleJobStatus</a>)
</p>
<p>
<p>ElasticStatus defines the status of elastic process</p>
</p>
<h3 id="batch.paddlepaddle.org/v1.Intranet">Intranet
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobSpec">PaddleJobSpec</a>)
</p>
<p>
</p>
<h3 id="batch.paddlepaddle.org/v1.PaddleJobMode">PaddleJobMode
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobStatus">PaddleJobStatus</a>)
</p>
<p>
<p>PaddleJobMode defines the avaiable mode of a job</p>
</p>
<h3 id="batch.paddlepaddle.org/v1.PaddleJobPhase">PaddleJobPhase
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobStatus">PaddleJobStatus</a>)
</p>
<p>
<p>PaddleJobPhase defines the phase of the job.</p>
</p>
<h3 id="batch.paddlepaddle.org/v1.PaddleJobSpec">PaddleJobSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1.PaddleJob">PaddleJob</a>)
</p>
<p>
<p>PaddleJobSpec defines the desired state of PaddleJob</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>cleanPodPolicy</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.CleanPodPolicy">
CleanPodPolicy
</a>
</em>
</td>
<td>
<p>CleanPodPolicy defines whether to clean pod after job finished</p>
</td>
</tr>
<tr>
<td>
<code>schedulingPolicy</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.SchedulingPolicy">
SchedulingPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchedulingPolicy defines the policy related to scheduling, for volcano</p>
</td>
</tr>
<tr>
<td>
<code>intranet</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.Intranet">
Intranet
</a>
</em>
</td>
<td>
<p>Intranet defines the communication mode inter pods : PodIP, Service or Host</p>
</td>
</tr>
<tr>
<td>
<code>withGloo</code></br>
<em>
int
</em>
</td>
<td>
<p>WithGloo indicate whether enable gloo, 0/1/2 for disable/enable for worker/enable for server</p>
</td>
</tr>
<tr>
<td>
<code>sampleSetRef</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.SampleSetRef">
SampleSetRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SampleSetRef defines the sample data set used for training and its mount path in worker pods</p>
</td>
</tr>
<tr>
<td>
<code>ps</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.ResourceSpec">
ResourceSpec
</a>
</em>
</td>
<td>
<p>PS[erver] describes the spec of server base on pod template</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.ResourceSpec">
ResourceSpec
</a>
</em>
</td>
<td>
<p>Worker describes the spec of worker base on pod template</p>
</td>
</tr>
<tr>
<td>
<code>heter</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.ResourceSpec">
ResourceSpec
</a>
</em>
</td>
<td>
<p>Heter describes the spec of heter worker base on pod temlate</p>
</td>
</tr>
<tr>
<td>
<code>elastic</code></br>
<em>
int
</em>
</td>
<td>
<p>Elastic indicate the elastic level</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1.PaddleJobStatus">PaddleJobStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1.PaddleJob">PaddleJob</a>)
</p>
<p>
<p>PaddleJobStatus defines the observed state of PaddleJob</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>phase</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobPhase">
PaddleJobPhase
</a>
</em>
</td>
<td>
<p>The phase of PaddleJob.</p>
</td>
</tr>
<tr>
<td>
<code>mode</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobMode">
PaddleJobMode
</a>
</em>
</td>
<td>
<p>Mode indicates in which the PaddleJob run with : PS/Collective/Single
PS mode is enabled when ps is set
Single/Collective is enabled if ps is missing</p>
</td>
</tr>
<tr>
<td>
<code>ps</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.ResourceStatus">
ResourceStatus
</a>
</em>
</td>
<td>
<p>ResourceStatues of ps</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.ResourceStatus">
ResourceStatus
</a>
</em>
</td>
<td>
<p>ResourceStatues of worker</p>
</td>
</tr>
<tr>
<td>
<code>heter</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.ResourceStatus">
ResourceStatus
</a>
</em>
</td>
<td>
<p>ResourceStatues of worker</p>
</td>
</tr>
<tr>
<td>
<code>elastic</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1.ElasticStatus">
ElasticStatus
</a>
</em>
</td>
<td>
<p>Elastic</p>
</td>
</tr>
<tr>
<td>
<code>startTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>StartTime indicate when the job started</p>
</td>
</tr>
<tr>
<td>
<code>completionTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>CompletionTime indicate when the job completed/failed</p>
</td>
</tr>
<tr>
<td>
<code>observedGeneration</code></br>
<em>
int
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1.ResourceSpec">ResourceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobSpec">PaddleJobSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>replicas</code></br>
<em>
int
</em>
</td>
<td>
<p>Replicas replica</p>
</td>
</tr>
<tr>
<td>
<code>requests</code></br>
<em>
int
</em>
</td>
<td>
<p>Requests set the minimal replicas of server to be run</p>
</td>
</tr>
<tr>
<td>
<code>limits</code></br>
<em>
int
</em>
</td>
<td>
<p>Requests set the maximal replicas of server to be run, elastic is auto enbale if limits is set larger than 0</p>
</td>
</tr>
<tr>
<td>
<code>template</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#podtemplatespec-v1-core">
Kubernetes core/v1.PodTemplateSpec
</a>
</em>
</td>
<td>
<p>Template specifies the podspec of a server</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1.ResourceStatus">ResourceStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobStatus">PaddleJobStatus</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>pending</code></br>
<em>
int
</em>
</td>
<td>
<p>Pending</p>
</td>
</tr>
<tr>
<td>
<code>starting</code></br>
<em>
int
</em>
</td>
<td>
<p>Starting</p>
</td>
</tr>
<tr>
<td>
<code>running</code></br>
<em>
int
</em>
</td>
<td>
<p>Running</p>
</td>
</tr>
<tr>
<td>
<code>failed</code></br>
<em>
int
</em>
</td>
<td>
<p>Failed</p>
</td>
</tr>
<tr>
<td>
<code>succeeded</code></br>
<em>
int
</em>
</td>
<td>
<p>Success</p>
</td>
</tr>
<tr>
<td>
<code>unknown</code></br>
<em>
int
</em>
</td>
<td>
<p>Unknown</p>
</td>
</tr>
<tr>
<td>
<code>refs</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectreference-v1-core">
[]Kubernetes core/v1.ObjectReference
</a>
</em>
</td>
<td>
<p>A list of pointer to pods</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1.SampleSetRef">SampleSetRef
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobSpec">PaddleJobSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name of the SampleSet.</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace of the SampleSet.</p>
</td>
</tr>
<tr>
<td>
<code>mountPath</code></br>
<em>
string
</em>
</td>
<td>
<p>Path within the container at which the volume should be mounted.  Must not contain &lsquo;:&rsquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1.SchedulingPolicy">SchedulingPolicy
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1.PaddleJobSpec">PaddleJobSpec</a>)
</p>
<p>
<p>SchedulingPolicy embed schedule policy of volcano</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>minAvailable</code></br>
<em>
int32
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>queue</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>priorityClass</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>minResources</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<p>pointer may cause deepcopy error
api/v1/zz_generated.deepcopy.go:230:8: cannot use new(map[&ldquo;k8s.io/api/core/v1&rdquo;.ResourceName]resource.Quantity) (type *map[&ldquo;k8s.io/api/core/v1&rdquo;.ResourceName]resource.Quantity) as type *&ldquo;k8s.io/api/core/v1&rdquo;.ResourceList in assignment</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<h2 id="batch.paddlepaddle.org/v1alpha1">batch.paddlepaddle.org/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains SampleSet and SampleJob</p>
</p>
Resource Types:
<ul><li>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleJob">SampleJob</a>
</li><li>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSet">SampleSet</a>
</li></ul>
<h3 id="batch.paddlepaddle.org/v1alpha1.SampleJob">SampleJob
</h3>
<p>
<p>SampleJob is the Schema for the samplejobs API</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
batch.paddlepaddle.org/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>SampleJob</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleJobSpec">
SampleJobSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>type</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleJobType">
SampleJobType
</a>
</em>
</td>
<td>
<p>Job Type of SampleJob. One of the three types: <code>sync</code>, <code>warmup</code>, <code>rmr</code>, <code>clear</code></p>
</td>
</tr>
<tr>
<td>
<code>sampleSetRef</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>the information of reference SampleSet object</p>
</td>
</tr>
<tr>
<td>
<code>secretRef</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#secretreference-v1-core">
Kubernetes core/v1.SecretReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Used for sync job, if the source data storage requires additional authorization information, set this field.</p>
</td>
</tr>
<tr>
<td>
<code>terminate</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>terminate other jobs that already in event queue of runtime servers</p>
</td>
</tr>
<tr>
<td>
<code>JobOptions</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.JobOptions">
JobOptions
</a>
</em>
</td>
<td>
<p>
(Members of <code>JobOptions</code> are embedded into this type.)
</p>
<em>(Optional)</em>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleJobStatus">
SampleJobStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.SampleSet">SampleSet
</h3>
<p>
<p>SampleSet is the Schema for the SampleSets API</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
batch.paddlepaddle.org/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>SampleSet</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSetSpec">
SampleSetSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>partitions</code></br>
<em>
int32
</em>
</td>
<td>
<p>Partitions is the number of SampleSet partitions, partition means cache node.</p>
</td>
</tr>
<tr>
<td>
<code>source</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.Source">
Source
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Source describes the information of data source uri and secret name.
cannot update after data sync finish</p>
</td>
</tr>
<tr>
<td>
<code>secretRef</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#secretreference-v1-core">
Kubernetes core/v1.SecretReference
</a>
</em>
</td>
<td>
<p>SecretRef is reference to the authentication secret for source storage and cache engine.
cannot update after SampleSet phase is Bound</p>
</td>
</tr>
<tr>
<td>
<code>noSync</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>If the data is already in cache engine backend storage, can set NoSync as true to skip Syncing phase.
cannot update after data sync finish</p>
</td>
</tr>
<tr>
<td>
<code>csi</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.CSI">
CSI
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CSI defines the csi driver and mount options for supporting dataset.
Cannot update after SampleSet phase is Bound</p>
</td>
</tr>
<tr>
<td>
<code>cache</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.Cache">
Cache
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Cache options used by cache runtime engine
Cannot update after SampleSet phase is Bound</p>
</td>
</tr>
<tr>
<td>
<code>nodeAffinity</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#nodeaffinity-v1-core">
Kubernetes core/v1.NodeAffinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeAffinity defines constraints that limit what nodes this SampleSet can be cached to.
This field influences the scheduling of pods that use the cached dataset.
Cannot update after SampleSet phase is Bound</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s tolerations.
Cannot update after SampleSet phase is Bound</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSetStatus">
SampleSetStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.CSI">CSI
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSetSpec">SampleSetSpec</a>)
</p>
<p>
<p>CSI describes csi driver name and mount options to support cache data</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>driver</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.DriverName">
DriverName
</a>
</em>
</td>
<td>
<p>Name of cache runtime driver, now only support juicefs.</p>
</td>
</tr>
<tr>
<td>
<code>MountOptions</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.MountOptions">
MountOptions
</a>
</em>
</td>
<td>
<p>
(Members of <code>MountOptions</code> are embedded into this type.)
</p>
<em>(Optional)</em>
<p>Namespace of the runtime object</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.Cache">Cache
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSetSpec">SampleSetSpec</a>)
</p>
<p>
<p>Cache used to describe how cache data store</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>levels</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.CacheLevel">
[]CacheLevel
</a>
</em>
</td>
<td>
<p>configurations for multiple storage tier</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.CacheLevel">CacheLevel
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.Cache">Cache</a>)
</p>
<p>
<p>CacheLevel describes configurations a tier needs</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>mediumType</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.MediumType">
MediumType
</a>
</em>
</td>
<td>
<p>Medium Type of the tier. One of the three types: <code>MEM</code>, <code>SSD</code>, <code>HDD</code></p>
</td>
</tr>
<tr>
<td>
<code>path</code></br>
<em>
string
</em>
</td>
<td>
<p>directory paths of local cache, use colon to separate multiple paths
For example: &ldquo;/dev/shm/cache1:/dev/ssd/cache2:/mnt/cache3&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>cacheSize</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>CacheSize size of cached objects in MiB
If multiple paths used for this, the cache size is total amount of cache objects in all paths</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.CacheStatus">CacheStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSetStatus">SampleSetStatus</a>)
</p>
<p>
<p>CacheStatus status of cache data</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>totalSize</code></br>
<em>
string
</em>
</td>
<td>
<p>TotalSize the total size of SampleSet data</p>
</td>
</tr>
<tr>
<td>
<code>totalFiles</code></br>
<em>
string
</em>
</td>
<td>
<p>TotalFiles the total file number of SampleSet data</p>
</td>
</tr>
<tr>
<td>
<code>cachedSize</code></br>
<em>
string
</em>
</td>
<td>
<p>CachedSize the total size of cached data in all nodes</p>
</td>
</tr>
<tr>
<td>
<code>diskSize</code></br>
<em>
string
</em>
</td>
<td>
<p>DiskSize disk space on file system containing cache data</p>
</td>
</tr>
<tr>
<td>
<code>diskUsed</code></br>
<em>
string
</em>
</td>
<td>
<p>DiskUsed disk space already been used, display by command df</p>
</td>
</tr>
<tr>
<td>
<code>diskAvail</code></br>
<em>
string
</em>
</td>
<td>
<p>DiskAvail disk space available on file system, display by command df</p>
</td>
</tr>
<tr>
<td>
<code>diskUsageRate</code></br>
<em>
string
</em>
</td>
<td>
<p>DiskUsageRate disk space usage rate display by command df</p>
</td>
</tr>
<tr>
<td>
<code>errorMassage</code></br>
<em>
string
</em>
</td>
<td>
<p>ErrorMassage error massages collected when executing related command</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.ClearJobOptions">ClearJobOptions
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.JobOptions">JobOptions</a>)
</p>
<p>
<p>ClearJobOptions the options for clear cache from local host</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>paths</code></br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.CronJobStatus">CronJobStatus
</h3>
<p>
<p>CronJobStatus represents the current state of a cron job.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>active</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectreference-v1-core">
[]Kubernetes core/v1.ObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>A list of pointers to currently running jobs.</p>
</td>
</tr>
<tr>
<td>
<code>lastScheduleTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Information when was the last time the job was successfully scheduled.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.DriverName">DriverName
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.CSI">CSI</a>)
</p>
<p>
<p>DriverName specified the name of csi driver</p>
</p>
<h3 id="batch.paddlepaddle.org/v1alpha1.JobOptions">JobOptions
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleJobSpec">SampleJobSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>syncOptions</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.SyncJobOptions">
SyncJobOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>sync job options</p>
</td>
</tr>
<tr>
<td>
<code>warmupOptions</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.WarmupJobOptions">
WarmupJobOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>warmup job options</p>
</td>
</tr>
<tr>
<td>
<code>rmrOptions</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.RmrJobOptions">
RmrJobOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>rmr job options</p>
</td>
</tr>
<tr>
<td>
<code>clearOptions</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.ClearJobOptions">
ClearJobOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>clear job options</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.JobsName">JobsName
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSetStatus">SampleSetStatus</a>)
</p>
<p>
<p>JobsName record the name of jobs triggered by SampleSet controller, it should store and load atomically.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>syncJobName</code></br>
<em>
k8s.io/apimachinery/pkg/types.UID
</em>
</td>
<td>
<p>the name of sync data job, used by controller to request runtime server for get job information.</p>
</td>
</tr>
<tr>
<td>
<code>warmupJobName</code></br>
<em>
k8s.io/apimachinery/pkg/types.UID
</em>
</td>
<td>
<p>record the name of the last done successfully sync job name</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.JuiceFSMountOptions">JuiceFSMountOptions
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.MountOptions">MountOptions</a>)
</p>
<p>
<p>JuiceFSMountOptions describes the JuiceFS mount options which user can set
All the mount options is list in <a href="https://github.com/juicedata/juicefs/blob/main/docs/en/command_reference.md">https://github.com/juicedata/juicefs/blob/main/docs/en/command_reference.md</a></p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metrics</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>address to export metrics (default: &ldquo;127.0.0.1:9567&rdquo;)</p>
</td>
</tr>
<tr>
<td>
<code>attr-cache</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>attributes cache timeout in seconds (default: 1)</p>
</td>
</tr>
<tr>
<td>
<code>entry-cache</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>file entry cache timeout in seconds (default: 1)</p>
</td>
</tr>
<tr>
<td>
<code>dir-entry-cache</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>dir entry cache timeout in seconds (default: 1)</p>
</td>
</tr>
<tr>
<td>
<code>enable-xattr</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>enable extended attributes (xattr) (default: false)</p>
</td>
</tr>
<tr>
<td>
<code>get-timeout</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>the max number of seconds to download an object (default: 60)</p>
</td>
</tr>
<tr>
<td>
<code>put-timeout</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>the max number of seconds to upload an object (default: 60)</p>
</td>
</tr>
<tr>
<td>
<code>io-retries</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>number of retries after network failure (default: 30)</p>
</td>
</tr>
<tr>
<td>
<code>max-uploads</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>number of connections to upload (default: 20)</p>
</td>
</tr>
<tr>
<td>
<code>buffer-size</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>total read/write buffering in MB (default: 300)</p>
</td>
</tr>
<tr>
<td>
<code>prefetch</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>prefetch N blocks in parallel (default: 1)</p>
</td>
</tr>
<tr>
<td>
<code>writeback</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>upload objects in background (default: false)</p>
</td>
</tr>
<tr>
<td>
<code>cache-dir</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>directory paths of local cache, use colon to separate multiple paths</p>
</td>
</tr>
<tr>
<td>
<code>cache-size</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>size of cached objects in MiB (default: 1024)</p>
</td>
</tr>
<tr>
<td>
<code>free-space-ratio</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>min free space (ratio) (default: 0.1)
float64 is not supported <a href="https://github.com/kubernetes-sigs/controller-tools/issues/245">https://github.com/kubernetes-sigs/controller-tools/issues/245</a></p>
</td>
</tr>
<tr>
<td>
<code>cache-partial-only</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>cache only random/small read (default: false)</p>
</td>
</tr>
<tr>
<td>
<code>open-cache</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>open files cache timeout in seconds (0 means disable this feature) (default: 0)</p>
</td>
</tr>
<tr>
<td>
<code>sub-dir</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>mount a sub-directory as root</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.JuiceFSSyncOptions">JuiceFSSyncOptions
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SyncJobOptions">SyncJobOptions</a>)
</p>
<p>
<p>JuiceFSSyncOptions describes the JuiceFS sync options which user can set by SampleSet</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>start</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>the first KEY to sync</p>
</td>
</tr>
<tr>
<td>
<code>end</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>the last KEY to sync</p>
</td>
</tr>
<tr>
<td>
<code>threads</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>number of concurrent threads (default: 10)</p>
</td>
</tr>
<tr>
<td>
<code>http-port</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>HTTP PORT to listen to (default: 6070)</p>
</td>
</tr>
<tr>
<td>
<code>update</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>update existing file if the source is newer (default: false)</p>
</td>
</tr>
<tr>
<td>
<code>force-update</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>always update existing file (default: false)</p>
</td>
</tr>
<tr>
<td>
<code>perms</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>preserve permissions (default: false)</p>
</td>
</tr>
<tr>
<td>
<code>dirs</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Sync directories or holders (default: false)</p>
</td>
</tr>
<tr>
<td>
<code>dry</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Don&rsquo;t copy file (default: false)</p>
</td>
</tr>
<tr>
<td>
<code>delete-src</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>delete objects from source after synced (default: false)</p>
</td>
</tr>
<tr>
<td>
<code>delete-dst</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>delete extraneous objects from destination (default: false)</p>
</td>
</tr>
<tr>
<td>
<code>exclude</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>exclude keys containing PATTERN (POSIX regular expressions)</p>
</td>
</tr>
<tr>
<td>
<code>include</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>only include keys containing PATTERN (POSIX regular expressions)</p>
</td>
</tr>
<tr>
<td>
<code>manager</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>manager address</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>hosts (seperated by comma) to launch worker</p>
</td>
</tr>
<tr>
<td>
<code>bwlimit</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>limit bandwidth in Mbps (0 means unlimited) (default: 0)</p>
</td>
</tr>
<tr>
<td>
<code>no-https</code></br>
<em>
bool
</em>
</td>
<td>
<p>do not use HTTPS (default: false)</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.JuiceFSWarmupOptions">JuiceFSWarmupOptions
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.WarmupJobOptions">WarmupJobOptions</a>)
</p>
<p>
<p>JuiceFSWarmupOptions describes the JuiceFS warmup options which user can set by SampleSet</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>file</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>the relative path of file that containing a list of data file paths</p>
</td>
</tr>
<tr>
<td>
<code>threads</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>number of concurrent workers (default: 50)</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.MediumType">MediumType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.CacheLevel">CacheLevel</a>)
</p>
<p>
<p>MediumType store medium type</p>
</p>
<h3 id="batch.paddlepaddle.org/v1alpha1.MountOptions">MountOptions
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.CSI">CSI</a>)
</p>
<p>
<p>MountOptions the mount options for csi drivers</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>juiceFSMountOptions</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.JuiceFSMountOptions">
JuiceFSMountOptions
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>JuiceFSMountOptions juicefs mount command options</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.RmrJobOptions">RmrJobOptions
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.JobOptions">JobOptions</a>)
</p>
<p>
<p>RmrJobOptions the options for remove data from cache engine</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>paths</code></br>
<em>
[]string
</em>
</td>
<td>
<p>Paths should be relative path from source directory, and without prefix.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.RuntimeStatus">RuntimeStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSetStatus">SampleSetStatus</a>)
</p>
<p>
<p>RuntimeStatus status of runtime StatefulSet</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>runtimeReady</code></br>
<em>
string
</em>
</td>
<td>
<p>RuntimeReady is use to display SampleSet Runtime pods status, format like {ReadyReplicas}/{SpecReplicas}.</p>
</td>
</tr>
<tr>
<td>
<code>specReplicas</code></br>
<em>
int32
</em>
</td>
<td>
<p>SpecReplicas is the number of Pods should be created by Runtime StatefulSet.</p>
</td>
</tr>
<tr>
<td>
<code>readyReplicas</code></br>
<em>
int32
</em>
</td>
<td>
<p>ReadyReplicas is the number of Pods created by the Runtime StatefulSet that have a Ready Condition.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.SampleJobPhase">SampleJobPhase
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleJobStatus">SampleJobStatus</a>)
</p>
<p>
</p>
<h3 id="batch.paddlepaddle.org/v1alpha1.SampleJobSpec">SampleJobSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleJob">SampleJob</a>)
</p>
<p>
<p>SampleJobSpec defines the desired state of SampleJob</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleJobType">
SampleJobType
</a>
</em>
</td>
<td>
<p>Job Type of SampleJob. One of the three types: <code>sync</code>, <code>warmup</code>, <code>rmr</code>, <code>clear</code></p>
</td>
</tr>
<tr>
<td>
<code>sampleSetRef</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>the information of reference SampleSet object</p>
</td>
</tr>
<tr>
<td>
<code>secretRef</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#secretreference-v1-core">
Kubernetes core/v1.SecretReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Used for sync job, if the source data storage requires additional authorization information, set this field.</p>
</td>
</tr>
<tr>
<td>
<code>terminate</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>terminate other jobs that already in event queue of runtime servers</p>
</td>
</tr>
<tr>
<td>
<code>JobOptions</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.JobOptions">
JobOptions
</a>
</em>
</td>
<td>
<p>
(Members of <code>JobOptions</code> are embedded into this type.)
</p>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.SampleJobStatus">SampleJobStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleJob">SampleJob</a>)
</p>
<p>
<p>SampleJobStatus defines the observed state of SampleJob</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>phase</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleJobPhase">
SampleJobPhase
</a>
</em>
</td>
<td>
<p>The phase of SampleJob is a simple, high-level summary of where the SampleJob is in its lifecycle.</p>
</td>
</tr>
<tr>
<td>
<code>jobName</code></br>
<em>
k8s.io/apimachinery/pkg/types.UID
</em>
</td>
<td>
<p>the uuid for a job, used by controller to post and get the job options and requests.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.SampleJobType">SampleJobType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleJobSpec">SampleJobSpec</a>)
</p>
<p>
</p>
<h3 id="batch.paddlepaddle.org/v1alpha1.SampleSetPhase">SampleSetPhase
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSetStatus">SampleSetStatus</a>)
</p>
<p>
<p>SampleSetPhase indicates whether the loading is behaving</p>
</p>
<h3 id="batch.paddlepaddle.org/v1alpha1.SampleSetSpec">SampleSetSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSet">SampleSet</a>)
</p>
<p>
<p>SampleSetSpec defines the desired state of SampleSet</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>partitions</code></br>
<em>
int32
</em>
</td>
<td>
<p>Partitions is the number of SampleSet partitions, partition means cache node.</p>
</td>
</tr>
<tr>
<td>
<code>source</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.Source">
Source
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Source describes the information of data source uri and secret name.
cannot update after data sync finish</p>
</td>
</tr>
<tr>
<td>
<code>secretRef</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#secretreference-v1-core">
Kubernetes core/v1.SecretReference
</a>
</em>
</td>
<td>
<p>SecretRef is reference to the authentication secret for source storage and cache engine.
cannot update after SampleSet phase is Bound</p>
</td>
</tr>
<tr>
<td>
<code>noSync</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>If the data is already in cache engine backend storage, can set NoSync as true to skip Syncing phase.
cannot update after data sync finish</p>
</td>
</tr>
<tr>
<td>
<code>csi</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.CSI">
CSI
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CSI defines the csi driver and mount options for supporting dataset.
Cannot update after SampleSet phase is Bound</p>
</td>
</tr>
<tr>
<td>
<code>cache</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.Cache">
Cache
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Cache options used by cache runtime engine
Cannot update after SampleSet phase is Bound</p>
</td>
</tr>
<tr>
<td>
<code>nodeAffinity</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#nodeaffinity-v1-core">
Kubernetes core/v1.NodeAffinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeAffinity defines constraints that limit what nodes this SampleSet can be cached to.
This field influences the scheduling of pods that use the cached dataset.
Cannot update after SampleSet phase is Bound</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s tolerations.
Cannot update after SampleSet phase is Bound</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.SampleSetStatus">SampleSetStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSet">SampleSet</a>)
</p>
<p>
<p>SampleSetStatus defines the observed state of SampleSet</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>phase</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSetPhase">
SampleSetPhase
</a>
</em>
</td>
<td>
<p>Dataset Phase. One of the four phases: <code>None</code>, <code>Bound</code>, <code>NotBound</code> and <code>Failed</code></p>
</td>
</tr>
<tr>
<td>
<code>cacheStatus</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.CacheStatus">
CacheStatus
</a>
</em>
</td>
<td>
<p>CacheStatus the status of cache data in cluster</p>
</td>
</tr>
<tr>
<td>
<code>runtimeStatus</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.RuntimeStatus">
RuntimeStatus
</a>
</em>
</td>
<td>
<p>RuntimeStatus the status of runtime StatefulSet pods</p>
</td>
</tr>
<tr>
<td>
<code>jobsName</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.JobsName">
JobsName
</a>
</em>
</td>
<td>
<p>recorde the name of jobs, all names is generated by uuid</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.SampleStrategy">SampleStrategy
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.WarmupJobOptions">WarmupJobOptions</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>strategyName</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.Source">Source
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleSetSpec">SampleSetSpec</a>)
</p>
<p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>uri</code></br>
<em>
string
</em>
</td>
<td>
<p>URI should be in the following format: [NAME://]BUCKET[.ENDPOINT][/PREFIX]
Cannot be updated after SampleSet sync data to cache engine
More info: <a href="https://github.com/juicedata/juicesync">https://github.com/juicedata/juicesync</a></p>
</td>
</tr>
<tr>
<td>
<code>secretRef</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#secretreference-v1-core">
Kubernetes core/v1.SecretReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If the remote storage requires additional authorization information, set this secret reference</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.SyncJobOptions">SyncJobOptions
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.JobOptions">JobOptions</a>)
</p>
<p>
<p>SyncJobOptions the options for sync data to cache engine</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>source</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>data source that need sync to cache engine, the format of it should be
[NAME://]BUCKET[.ENDPOINT][/PREFIX]</p>
</td>
</tr>
<tr>
<td>
<code>destination</code></br>
<em>
string
</em>
</td>
<td>
<p>the relative path in mount volume for data sync to, eg: /train</p>
</td>
</tr>
<tr>
<td>
<code>JuiceFSSyncOptions</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.JuiceFSSyncOptions">
JuiceFSSyncOptions
</a>
</em>
</td>
<td>
<p>
(Members of <code>JuiceFSSyncOptions</code> are embedded into this type.)
</p>
<em>(Optional)</em>
<p>JuiceFS sync command options</p>
</td>
</tr>
</tbody>
</table>
<h3 id="batch.paddlepaddle.org/v1alpha1.WarmupJobOptions">WarmupJobOptions
</h3>
<p>
(<em>Appears on:</em>
<a href="#batch.paddlepaddle.org/v1alpha1.JobOptions">JobOptions</a>)
</p>
<p>
<p>WarmupJobOptions the options for warmup date to local host</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>paths</code></br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>A list of paths need to build cache</p>
</td>
</tr>
<tr>
<td>
<code>partitions</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>the partitions of cache data, same as SampleSet spec.partitions</p>
</td>
</tr>
<tr>
<td>
<code>Strategy</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.SampleStrategy">
SampleStrategy
</a>
</em>
</td>
<td>
<p>
(Members of <code>Strategy</code> are embedded into this type.)
</p>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>JuiceFSWarmupOptions</code></br>
<em>
<a href="#batch.paddlepaddle.org/v1alpha1.JuiceFSWarmupOptions">
JuiceFSWarmupOptions
</a>
</em>
</td>
<td>
<p>
(Members of <code>JuiceFSWarmupOptions</code> are embedded into this type.)
</p>
<em>(Optional)</em>
<p>JuiceFS warmup command options</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>36cea5b</code>.
</em></p>
