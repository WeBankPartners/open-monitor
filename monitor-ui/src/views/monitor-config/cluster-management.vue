<template>
  <div class="main-content">
    <template v-for="cluster in clusterList">
      <Card style="width:20%;display:inline-block;margin:16px;" :key="cluster.id">
        <p slot="title">
          {{cluster.display_name}}
        </p>
        <a href="#" slot="extra" v-if="cluster.id !=='default'" @click.prevent="editF(cluster)">
          {{$t('button.edit')}}
        </a>
        <a href="#" slot="extra" v-if="cluster.id !=='default'" @click.prevent="syncMonitor(cluster)" style="color:#19be6b;">
          {{$t('button.syncMonitor')}}
        </a>
        <a href="#" slot="extra" v-if="cluster.id !=='default'" @click.prevent="deleteF(cluster)" style="color:red">
          {{$t('button.remove')}}
        </a>
        <ul>
          <li class="cluster-li">
            <div class="cluster-li-title">
              {{$t('id')}}:
            </div>
            <div class="cluster-li-value">
              {{cluster.id}}
            </div>
          </li>
          <li class="cluster-li">
            <div class="cluster-li-title">
              {{$t('m_prom_address')}}:
            </div>
            <div class="cluster-li-value">
              {{cluster.prom_address}}
            </div>
          </li>
          <li class="cluster-li">
            <div class="cluster-li-title">
            {{$t('m_remote_agent_address')}}:
            </div>
            <div class="cluster-li-value">
              {{cluster.remote_agent_address}}
            </div>
          </li>
        </ul>
      </Card>
    </template>
    <Card style="width:20%;display:inline-block;margin:16px;vertical-align: bottom;">
      <p slot="title">
        {{$t('button.add')}}
      </p>
      <div style="margin:8px;text-align:center;height:56px">
        <Icon @click="addItem" type="md-add-circle" :size=25 style="cursor:pointer" :color="'#2d8cf0'" />
      </div>
    </Card>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
    <Modal
    v-model="isShowWarning"
      :title="$t('delConfirm.title')"
      @on-ok="ok"
      @on-cancel="cancel">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('delConfirm.tip')}}</p>
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
          modalId: 'add_edit_Modal',
          modalTitle: 'm_cluster_management',
          isAdd: true,
          config: [
            {label: 'field.id', value: 'id', placeholder: 'tips.inputRequired', v_validate: 'required:true|min:2|max:60', disabled: true, type: 'text'},
            {label: 'field.displayName', value: 'display_name', placeholder: 'tips.inputRequired', v_validate: 'required:true|min:2|max:60', disabled: false, type: 'text'},
            {label: 'm_prom_address', value: 'prom_address', placeholder: '', disabled: false, type: 'text'},
            {label: 'm_remote_agent_address', value: 'remote_agent_address', placeholder: '', disabled: false, type: 'text'}
          ],
          addRow: { // [通用]-保存用户新增、编辑时数据
            id: '',
            display_name: '',
            prom_address: '',
            remote_agent_address: '',
          }
        },
        modelTip: {
          key: 'id',
          value: null
        },
        id: '',
        selectedData: ''
      }
    },
    mounted () {
      this.getClusterList()
    },
    methods: {
      getClusterList () {
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.cluster, '', (res) => {
          this.clusterList = res
        })
      },
      ok () {
        this.delF(this.selectedData)
      },
      deleteF (item) {
        this.isShowWarning = true
        this.selectedData = item
      },
      delF (rowData) {
        let params = {id: rowData.id}
        this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', this.$root.apiCenter.cluster, params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.getClusterList()
        })
      },
      cancel () {
        this.isShowWarning = false
      },
      addItem () {
        this.$root.JQ('#add_edit_Modal').modal('show')
      },
      addPost () {
        let params= JSON.parse(JSON.stringify(this.modelConfig.addRow))
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.cluster, params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.getClusterList()
        })
      },
      editPost () {
        let params= JSON.parse(JSON.stringify(this.modelConfig.addRow))
        this.$root.$httpRequestEntrance.httpRequestEntrance('PUT', this.$root.apiCenter.cluster, params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.getClusterList()
        })
      },
      syncMonitor (rowData) {
        let params = {
          cluster: rowData.id
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance('get', this.$root.apiCenter.cluster, params, () => {
          this.$Message.success(this.$t('tips.success'))
        })
      }, 
      async editF (rowData) {
        this.modelConfig.isAdd = false
        this.modelTip.value = rowData[this.modelTip.key]
        this.id = rowData.id
        this.modelConfig.addRow = {
          ...rowData
        }
        this.$root.JQ('#add_edit_Modal').modal('show')
      },
    },
    components: {
    }
  }
</script>

<style lang="less" scoped>
.cluster-li {
  margin: 8px;
  list-style: none;
}
.cluster-li-title {
  width: 40%;
  display: inline-block;
  font-size: 14px;
  font-weight: 500;
}
.cluster-li-value {
  width: 40%;
  display: inline-block;
}
</style>
