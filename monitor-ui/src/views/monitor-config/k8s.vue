<template>
  <div class=" ">
    <template v-for="cluster in clusterList">
      <Card style="width:350px;display:inline-block;margin:16px;" :key="cluster.id">
        <p slot="title">
          {{cluster.cluster_name.split(':')[0]}}
        </p>
        <a href="#" slot="extra" @click.prevent="editCluster(cluster)">
          {{$t('button.edit')}}
        </a>
        <a href="#" slot="extra" @click.prevent="deleteCluster({id: cluster.id})" style="color:red">
          {{$t('button.remove')}}
        </a>
        <ul>
          <li style="margin:8px;">
            <h5 style="width:30%;display:inline-block;text">
            {{$t('field.ip')}}:
            </h5>
            <div style="width:30%;display:inline-block;">
              {{cluster.api_server.split(':')[0]}}
            </div>
          </li>
          <li style="margin:8px;">
            <h5 style="width:30%;display:inline-block;">
              {{$t('field.port')}}:
            </h5>
            <div style="width:30%;display:inline-block;">
              {{cluster.api_server.split(':')[1]}}
            </div>
          </li>
        </ul>
      </Card>
    </template>
    <Card style="width:350px;display:inline-block;margin:16px;vertical-align: bottom;">
      <p slot="title">
        {{$t('button.add')}}
      </p>
      <div style="margin:14px;text-align:center">
        <Icon @click="addCluster" type="md-add-circle" :size=40 style="cursor:pointer" :color="'#2d8cf0'" />
      </div>
    </Card>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
    <Modal v-model="isShowWarning" title="Delete confirmation" @on-ok="ok" @on-cancel="cancel">
    <div class="modal-body" style="padding:30px">
      <div style="text-align:center">
        <p style="color: red">Will you delete it?</p>
      </div>
    </div>
  </Modal>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
      clusterList: [],
      modelConfig: {
        modalId: 'cluster_Modal',
        modalTitle: 'k8s',
        isAdd: true,
        config: [
            {
            label: 'clusterName',
            value: 'cluster_name',
            placeholder: 'tips.required',
            v_validate: 'required:true',
            disabled: false,
            type: 'text'
          },
          {
            label: 'field.ip',
            value: 'ip',
            placeholder: 'tips.required',
            v_validate: 'required:true|isIP',
            disabled: false,
            type: 'text'
          },
          {
            label: 'field.port',
            value: 'port',
            placeholder: 'tips.required',
            v_validate: 'required:true',
            disabled: false,
            type: 'text'
          },
          {
            label: 'token',
            value: 'token',
            placeholder: 'tips.required',
            v_validate: 'required:true',
            disabled: false,
            type: 'textarea',
            hide: 'edit'
          }
        ],
        addRow: {
          cluster_name: null,
          ip: null,
          port: null,
          token: null
        }
      },
      selectedData: null,
      modelTip: {
        key: 'cluster_name',
        value: null
      }
    }
  },
  mounted () {
    this.getClusterList()
  },
  methods: {
    addCluster () {
      this.modelConfig.isAdd = true
      this.$root.JQ('#cluster_Modal').modal('show')
    },
    addPost () {
      console.log(this.modelConfig.addRow)
      this.modelConfig.addRow.api_server = this.modelConfig.addRow.ip + ':' + this.modelConfig.addRow.port
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/agent/kubernetes/cluster/add', this.modelConfig.addRow, () => {
        this.$root.JQ('#cluster_Modal').modal('hide')
        this.getClusterList()
      })
    },
    getClusterList () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/agent/kubernetes/cluster/list', '', (responseData) => {
        this.clusterList = responseData
      })
    },
    deleteCluster (params) {
      this.selectedData = params
      this.isShowWarning = true
    },
    ok() {
      this.delF()
    },
    cancel() {
      this.isShowWarning = false
    },
    delF() {
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/agent/kubernetes/cluster/delete', this.selectedData, () => {
        this.isShowWarning = false
        this.getClusterList()
      })
    },
    editCluster (cluster) {
      this.modelTip.value = cluster.cluster_name
      this.modelConfig.isAdd = false
      this.modelConfig.addRow.id = cluster.id
      this.modelConfig.addRow.cluster_name = cluster.cluster_name
      this.modelConfig.addRow.ip = cluster.api_server.split(':')[0]
      this.modelConfig.addRow.port = cluster.api_server.split(':')[1]
      this.modelConfig.addRow.token = cluster.token
      this.$root.JQ('#cluster_Modal').modal('show')
    },
    editPost () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/agent/kubernetes/cluster/update', this.modelConfig.addRow, () => {
        this.$root.JQ('#cluster_Modal').modal('hide')
        this.getClusterList()
      })
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
.ivu-card-head {
  background: red;
}
</style>
