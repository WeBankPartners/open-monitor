<template>
  <div class="log-samples">
    <Button icon="ios-document-outline" @click="open">{{$t('m_k8s_log_samples')}}</Button>
    <Drawer v-model="visible" :width="85" :title="$t('m_k8s_log_samples')" @on-visible-change="visibleChanged">
      <p class="sample-context">{{object.clusterId}} / {{object.namespace}} / {{object.podName}} / {{object.container}}</p>
      <p class="sample-context">{{new Date(range.start * 1000).toLocaleString()}} — {{new Date(range.end * 1000).toLocaleString()}}</p>
      <Alert>{{$t('m_k8s_log_samples_help')}}</Alert>
      <Button :loading="loading" @click="load">{{$t('m_refresh')}}</Button>
      <Alert v-if="failed" type="error" class="sample-message">{{$t('m_k8s_log_samples_failed')}}</Alert>
      <Alert v-else-if="!loading && !rows.length" class="sample-message">{{$t('m_k8s_log_samples_empty')}}</Alert>
      <Alert v-if="limited" type="warning" class="sample-message">{{$t('m_k8s_log_samples_limited', {limit})}}</Alert>
      <div class="sample-list">
        <Spin v-if="loading" fix />
        <Card v-for="(sample, index) in pageRows" :key="index" class="sample-card">
          <div class="sample-heading">
            <strong>{{new Date(sample.observedAt * 1000).toLocaleString()}}</strong>
            <span>{{$t('m_k8s_rule_id')}}: {{sample.labels.rule}} / {{sample.labels.rule_version}}</span>
            <Button size="small" @click="copySample(sample)">{{$t('m_copy')}}</Button>
          </div>
          <p class="sample-location">{{$t('m_k8s_log_sample_file')}}: {{sample.labels.file}} · {{$t('m_k8s_log_sample_file_id')}}: {{sample.labels.file_id}} · {{$t('m_k8s_log_sample_offset')}}: {{sample.labels.byte_offset}}</p>
          <pre class="sample-preview">{{sample.labels.preview}}</pre>
          <Tag v-if="sample.labels.truncated === 'true'" color="orange">{{$t('m_k8s_log_sample_truncated')}}</Tag>
          <Collapse>
            <Panel name="labels">
              {{$t('m_k8s_log_sample_labels')}}
              <dl slot="content" class="sample-labels">
                <div v-for="(value, key) in sample.labels" v-show="key !== 'preview'" :key="key"><dt>{{key}}</dt><dd>{{value}}</dd></div>
              </dl>
            </Panel>
          </Collapse>
        </Card>
      </div>
      <Page v-if="rows.length" :total="rows.length" :current="page" :page-size="5" show-total @on-change="page = $event" />
    </Drawer>
  </div>
</template>
<script>
export default {
  props: {
    object: Object,
    range: Object,
    refreshToken: Number
  },
  data() {
    return {
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
      visible: false,
      loading: false,
      failed: false,
      limited: false,
      limit: 200,
      rows: [],
      page: 1
    }
  },
  computed: {
    queryKey() {
      const o = this.object
      return JSON.stringify([o.clusterId,o.namespace,o.workloadType,o.workloadName,o.podName,o.container,this.range.start,this.range.end,this.refreshToken])
    },
    pageRows() { return this.rows.slice((this.page-1)*5, this.page*5) }
  },
  watch: {
    queryKey() { if (this.visible) {this.load()} }
  },
  beforeDestroy() {
    this._destroyed = true
    this.visibleChanged(false)
  },
  methods: {
    open() {
      this.visible = true
      this.load()
    },
    visibleChanged(visible) {
      if (!visible) {
        this._generation = (this._generation || 0) + 1
        clearTimeout(this._deadline)
        this.loading = false
      }
    },
    load() {
      if (!this.visible || this._destroyed) {return}
      const generation = this._generation = (this._generation || 0) + 1
      clearTimeout(this._deadline)
      this.rows = []
      this.page = 1
      this.loading = true
      this.failed = false
      this.limited = false
      let finished = false
      const finish = (data, failed) => {
        if (this._generation !== generation || finished) {return}
        finished = true
        clearTimeout(this._deadline)
        this.loading = false
        this.failed = failed
        this.rows = failed ? [] : data.samples || []
        this.limited = !failed && data.limited === true
        this.limit = !failed && data.limit || 200
      }
      this._deadline = setTimeout(() => finish(null, true), 30000)
      const {
        clusterId, namespace, workloadType, workloadName, podName, container
      } = this.object
      this.request('POST', this.apiCenter.k8sLogSamples, {
        clusterId,
        namespace,
        workloadType,
        workloadName,
        podName,
        container,
        start: this.range.start,
        end: this.range.end
      }, data => finish(data, false), {isNeedloading: false}, () => finish(null, true))
    },
    copySample(sample) {
      const input = document.createElement('textarea')
      input.value = JSON.stringify(sample, null, 2)
      input.style.position = 'fixed'
      input.style.top = '-9999px'
      document.body.appendChild(input)
      input.select()
      try {
        if (!document.execCommand('copy')) {throw new Error('Copy failed')}
        this.$Message.success(this.$t('m_copied_to_clipboard'))
      } catch (err) {
        this.$Message.error(this.$t('m_k8s_log_sample_copy_failed'))
      } finally {
        input.remove()
      }
    }
  }
}
</script>
<style scoped>
.log-samples { margin:0 0 12px; }
.sample-context, .sample-location { overflow-wrap:anywhere; margin-bottom:8px; }
.sample-message { margin-top:12px; }
.sample-list { position:relative; min-height:60px; margin-top:12px; }
.sample-card { margin-bottom:12px; }
.sample-heading { display:flex; flex-wrap:wrap; align-items:center; gap:12px; margin-bottom:8px; }
.sample-heading span { overflow-wrap:anywhere; }
.sample-preview { white-space:pre-wrap; overflow-wrap:anywhere; word-break:break-word; max-height:220px; overflow:auto; padding:12px; background:#f5f7f9; }
.sample-labels > div { display:flex; gap:12px; padding:4px 0; overflow-wrap:anywhere; }
.sample-labels dt { flex:0 0 140px; font-weight:normal; }
.sample-labels dd { min-width:0; margin:0; }
@media (max-width:600px) {
  .sample-labels > div { display:block; }
}
</style>
