<!--采集器-->
<template>
  <div class=" ">
    <!-- <Divider plain orientation="left">K8s Com</Divider> -->
    <Card :key="'k8s'" :bordered="true" :dis-hover="true">
      <p slot="title">K8S</p>
      <template v-for="cluster in clusterList">
        <Card style="width:20%;display:inline-block;margin:16px;" :key="'k8s_' + cluster.id">
          <p slot="title">
            {{cluster.cluster_name.split(':')[0]}}
          </p>
          <a href="#" slot="extra" @click.prevent="editCluster(cluster)">
            {{$t('m_button_edit')}}
          </a>
          <a href="#" slot="extra" @click.prevent="deleteCluster({id: cluster.id, name: cluster.cluster_name})" style="color:red">
            {{$t('m_button_remove')}}
          </a>
          <ul>
            <li style="margin:8px;list-style: none;">
              <div style="width:30%;display:inline-block;font-size:16px;font-weight: 500;">
                {{$t('m_field_ip')}}:
              </div>
              <div style="width:30%;display:inline-block;">
                {{cluster.api_server}}
              </div>
            </li>
          </ul>
        </Card>
      </template>
      <Card style="width:20%;display:inline-block;margin:16px;vertical-align: bottom;">
        <p slot="title">
          {{$t('m_button_add')}}
        </p>
        <div style="margin:8px;text-align:center">
          <Icon @click="addCluster" type="md-add-circle" :size=25 style="cursor:pointer" :color="'#2d8cf0'" />
        </div>
      </Card>
    </Card>
    <Card :key="'snmp'" :bordered="true" :dis-hover="true" style="margin-top: 10px">
      <p slot="title">SNMP</p>
      <template v-for="item in snmpList">
        <Card style="width:20%;display:inline-block;margin:16px;" :key="'snmp_' + item.id">
          <p slot="title">
            {{item.id}}
          </p>
          <a href="#" slot="extra" @click.prevent="editItem(item)">
            {{$t('m_button_edit')}}
          </a>
          <a href="#" slot="extra" @click.prevent="deleteItem({id: item.id})" style="color:red">
            {{$t('m_button_remove')}}
          </a>
          <ul>
            <li style="margin:8px;list-style: none;">
              <div style="width:30%;display:inline-block;font-size:16px;font-weight: 500;">
                {{$t('m_field_ip')}}:
              </div>
              <div style="width:30%;display:inline-block;">
                {{item.address}}
              </div>
            </li>
          </ul>
        </Card>
      </template>
      <Card style="width:20%;display:inline-block;margin:16px;vertical-align: bottom;">
        <p slot="title">
          {{$t('m_button_add')}}
        </p>
        <div style="margin:8px;text-align:center">
          <Icon @click="addItem" type="md-add-circle" :size=25 style="cursor:pointer" :color="'#2d8cf0'" />
        </div>
      </Card>
    </Card>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
    <ModalComponent :modelConfig="modelItemConfig"></ModalComponent>
    <Modal
      v-model="isShowWarning"
      :title="$t('m_delConfirm_title')"
      :mask-closable="false"
      @on-ok="ok"
      @on-cancel="cancel"
    >
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{selectedData.name || selectedData.id}}{{$t('m_delConfirm_tip')}}</p>
        </div>
      </div>
    </Modal>
  </div>
</template>

<script>
import { collectionInterval } from '@/assets/config/common-config'
export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
      clusterList: [],
      snmpList: [],
      modelConfig: {
        modalId: 'cluster_Modal',
        modalTitle: 'm_proxy_exporter',
        isAdd: true,
        config: [
          {
            label: 'm_clusterName',
            value: 'cluster_name',
            placeholder: 'm_tips_required',
            v_validate: 'required:true',
            disabled: false,
            type: 'text'
          },
          {
            label: 'm_field_ip',
            value: 'ip',
            placeholder: 'm_tips_required',
            v_validate: 'required:true|isIP',
            disabled: false,
            type: 'text'
          },
          {
            label: 'm_field_port',
            value: 'port',
            placeholder: 'm_tips_required',
            v_validate: 'required:true',
            disabled: false,
            type: 'text'
          },
          {
            label: 'm_token',
            value: 'token',
            placeholder: 'm_tips_required',
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
      modelItemConfig: {
        modalId: 'item_Modal',
        modalTitle: 'm_proxy_exporter',
        isAdd: true,
        saveFunc: 'saveItem',
        config: [
          {
            label: 'm_field_id',
            value: 'id',
            placeholder: 'm_tips_required',
            v_validate: 'required:true',
            disabled: true,
            type: 'text'
          },
          {
            label: 'm_field_address',
            value: 'address',
            placeholder: 'm_exporter_address_placeholder',
            v_validate: 'required:true',
            disabled: false,
            type: 'text'
          },
          {
            label: 'm_collection_interval',
            value: 'scrape_interval',
            option: 'scrape_interval',
            placeholder: 'm_tips_required',
            v_validate: 'required:true',
            disabled: false,
            type: 'select'
          },
          {
            label: 'm_modules',
            value: 'modules',
            placeholder: 'if_mib',
            v_validate: 'required:true',
            disabled: false,
            type: 'text'
          },
        ],
        addRow: {
          id: null,
          address: null,
          scrape_interval: null,
          modules: 'if_mib'
        },
        v_select_configs: {
          scrape_interval: collectionInterval
        },
      },
      selectedData: {},
      selectedDataType: '',
      modelTip: {
        key: 'cluster_name',
        value: null
      }
    }
  },
  mounted() {
    this.getClusterList()
    this.getSnmpList()
  },
  methods: {
    addCluster() {
      this.modelConfig.isAdd = true
      this.$root.JQ('#cluster_Modal').modal('show')
    },
    addPost() {
      this.modelConfig.addRow.api_server = this.modelConfig.addRow.ip + ':' + this.modelConfig.addRow.port
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/agent/kubernetes/cluster/add', this.modelConfig.addRow, () => {
        this.$root.JQ('#cluster_Modal').modal('hide')
        this.getClusterList()
      })
    },
    getClusterList() {
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/agent/kubernetes/cluster/list', '', responseData => {
        this.clusterList = responseData
      })
    },
    deleteCluster(params) {
      this.selectedData = params
      this.selectedDataType = 'k8s'
      this.isShowWarning = true
    },
    ok() {
      this.delF()
    },
    cancel() {
      this.isShowWarning = false
    },
    delF() {
      if (this.selectedDataType === 'k8s') {
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/agent/kubernetes/cluster/delete', this.selectedData, () => {
          this.isShowWarning = false
          this.getClusterList()
        })
      } else if (this.selectedDataType === 'snmp') {
        this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', '/monitor/api/v1/config/new/snmp', this.selectedData, () => {
          this.isShowWarning = false
          this.getSnmpList()
        })
      }
    },
    editCluster(cluster) {
      this.modelTip.value = cluster.cluster_name
      this.modelConfig.isAdd = false
      this.modelConfig.addRow.id = cluster.id
      this.modelConfig.addRow.cluster_name = cluster.cluster_name
      this.modelConfig.addRow.ip = cluster.api_server.split(':')[0]
      this.modelConfig.addRow.port = cluster.api_server.split(':')[1]
      this.modelConfig.addRow.token = cluster.token
      this.$root.JQ('#cluster_Modal').modal('show')
    },
    editPost() {
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/agent/kubernetes/cluster/update', this.modelConfig.addRow, () => {
        this.$root.JQ('#cluster_Modal').modal('hide')
        this.getClusterList()
      })
    },
    getSnmpList() {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/config/new/snmp', '', responseData => {
        this.snmpList = responseData
      })
    },
    addItem() {
      this.modelItemConfig.isAdd = true
      this.modelItemConfig.addRow.id = null
      this.modelItemConfig.addRow.address = null
      this.modelItemConfig.addRow.scrape_interval = 10
      this.modelItemConfig.addRow.modules = 'if_mib'
      this.$root.JQ('#item_Modal').modal('show')
    },
    editItem(item) {
      this.modelTip.value = item.id
      this.modelItemConfig.isAdd = false
      this.modelItemConfig.addRow.id = item.id
      this.modelItemConfig.addRow.address = item.address
      this.modelItemConfig.addRow.scrape_interval = item.scrape_interval
      this.modelItemConfig.addRow.modules = item.modules

      this.$root.JQ('#item_Modal').modal('show')
    },
    deleteItem(params) {
      this.selectedData = params
      this.selectedDataType = 'snmp'
      this.isShowWarning = true
    },
    saveItem() {
      if (this.modelItemConfig.isAdd) {
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/config/new/snmp', this.modelItemConfig.addRow, () => {
          this.$root.JQ('#item_Modal').modal('hide')
          this.getSnmpList()
        })
      } else {
        this.$root.$httpRequestEntrance.httpRequestEntrance('PUT', '/monitor/api/v1/config/new/snmp', this.modelItemConfig.addRow, () => {
          this.$root.JQ('#item_Modal').modal('hide')
          this.getSnmpList()
        })
      }
    }
  },
  components: {},
}
</script>

<style scoped lang="less">

</style>
