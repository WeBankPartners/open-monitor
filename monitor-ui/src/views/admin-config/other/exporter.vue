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
          <Icon @click="addCluster" type="md-add-circle" :size=25 style="cursor:pointer" :color="'#5384FF'" />
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
          <Icon @click="addItem" type="md-add-circle" :size=25 style="cursor:pointer" :color="'#5384FF'" />
        </div>
      </Card>
    </Card>
    <Card :key="'custom'" :bordered="true" :dis-hover="true" style="margin-top: 10px">
      <p slot="title">{{$t('m_custom_scrape')}}</p>
      <template v-for="item in customScrapeList">
        <Card style="width:20%;display:inline-block;margin:16px;" :key="'custom_' + item.guid">
          <p slot="title">
            {{item.name}}
          </p>
          <a href="#" slot="extra" @click.prevent="editCustomItem(item)">
            {{$t('m_button_edit')}}
          </a>
          <a href="#" slot="extra" @click.prevent="deleteCustomItem(item)" style="color:red">
            {{$t('m_button_remove')}}
          </a>
          <ul>
            <li style="margin:8px;list-style: none;">
              <div style="width:30%;display:inline-block;font-size:16px;font-weight: 500;">
                {{$t('m_job_name')}}:
              </div>
              <div style="width:65%;display:inline-block;word-break:break-all;">
                {{item.job_name}}
              </div>
            </li>
            <li style="margin:8px;list-style: none;">
              <div style="width:30%;display:inline-block;font-size:16px;font-weight: 500;">
                {{$t('m_enabled')}}:
              </div>
              <div style="width:65%;display:inline-block;">
                {{item.enabled === 1 ? $t('m_enabled') : $t('m_disable')}}
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
          <Icon @click="addCustomItem" type="md-add-circle" :size=25 style="cursor:pointer" :color="'#5384FF'" />
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
    <Modal
      v-model="customModalVisible"
      :title="$t('m_custom_scrape')"
      :mask-closable="false"
      :width="860"
      @on-cancel="closeCustomModal"
    >
      <div>
        <p style="margin-bottom: 12px;color:#808695;">{{$t('m_custom_scrape_tip')}}</p>
        <Form :label-width="120">
          <FormItem :label="$t('m_name')" required>
            <Input v-model="customForm.name" :placeholder="$t('m_tips_required')" />
          </FormItem>
          <FormItem :label="$t('m_enabled')">
            <i-switch v-model="customForm.enabled" :true-value="1" :false-value="0" />
          </FormItem>
          <FormItem :label="$t('m_custom_scrape_yaml')" required>
            <Input
              v-model="customForm.yaml_content"
              type="textarea"
              :rows="16"
              :placeholder="$t('m_custom_scrape_yaml_placeholder')"
              class="custom-scrape-yaml"
            />
          </FormItem>
        </Form>
      </div>
      <div slot="footer">
        <Button @click="formatCustomYaml">{{$t('m_button_format_indent')}}</Button>
        <Button @click="validateCustomYaml">{{$t('m_button_validate')}}</Button>
        <Button @click="closeCustomModal">{{$t('m_button_cancel')}}</Button>
        <Button type="primary" @click="saveCustomItem">{{$t('m_button_save')}}</Button>
      </div>
    </Modal>
  </div>
</template>

<script>
import { collectionInterval } from '@/assets/config/common-config'
import { formatScrapeYaml } from '@/assets/js/format-scrape-yaml'
export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
      customModalVisible: false,
      clusterList: [],
      snmpList: [],
      customScrapeList: [],
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
            disabled: true,
            type: 'text'
          },
          {
            label: 'm_field_port',
            value: 'port',
            placeholder: 'm_tips_required',
            v_validate: 'required:true',
            disabled: true,
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
      customForm: {
        isAdd: true,
        guid: '',
        name: '',
        yaml_content: '',
        enabled: 1
      },
      modelTip: {
        key: 'cluster_name',
        value: null
      },
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  mounted() {
    this.getClusterList()
    this.getSnmpList()
    this.getCustomScrapeList()
  },
  methods: {
    addCluster() {
      this.modelConfig.isAdd = true
      this.$root.JQ('#cluster_Modal').modal('show')
    },
    addPost() {
      this.modelConfig.addRow.api_server = this.modelConfig.addRow.ip + ':' + this.modelConfig.addRow.port
      this.request('POST', this.apiCenter.addKubernetesCluster, this.modelConfig.addRow, () => {
        this.$root.JQ('#cluster_Modal').modal('hide')
        this.getClusterList()
      })
    },
    getClusterList() {
      this.request('POST', this.apiCenter.kubernetesClusterList, '', responseData => {
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
        this.request('POST', this.apiCenter.deleteKubernetesCluster, this.selectedData, () => {
          this.isShowWarning = false
          this.getClusterList()
        })
      } else if (this.selectedDataType === 'snmp') {
        this.request('DELETE', this.apiCenter.newSnmpConfig, this.selectedData, () => {
          this.isShowWarning = false
          this.getSnmpList()
        })
      } else if (this.selectedDataType === 'custom') {
        this.request('DELETE', this.apiCenter.customScrapeConfig, { guid: this.selectedData.guid }, () => {
          this.isShowWarning = false
          this.getCustomScrapeList()
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
      this.request('POST', this.apiCenter.updateKubernetesCluster, this.modelConfig.addRow, () => {
        this.$root.JQ('#cluster_Modal').modal('hide')
        this.getClusterList()
      })
    },
    getSnmpList() {
      this.request('GET', this.apiCenter.newSnmpConfig, '', responseData => {
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
        this.request('POST', this.apiCenter.newSnmpConfig, this.modelItemConfig.addRow, () => {
          this.$root.JQ('#item_Modal').modal('hide')
          this.getSnmpList()
        })
      } else {
        this.request('PUT', this.apiCenter.newSnmpConfig, this.modelItemConfig.addRow, () => {
          this.$root.JQ('#item_Modal').modal('hide')
          this.getSnmpList()
        })
      }
    },
    getCustomScrapeList() {
      this.request('GET', this.apiCenter.customScrapeConfig, '', responseData => {
        this.customScrapeList = responseData || []
      })
    },
    addCustomItem() {
      this.customForm = {
        isAdd: true,
        guid: '',
        name: '',
        yaml_content: '',
        enabled: 1
      }
      this.customModalVisible = true
    },
    editCustomItem(item) {
      this.customForm = {
        isAdd: false,
        guid: item.guid,
        name: item.name,
        yaml_content: item.yaml_content,
        enabled: item.enabled === 0 ? 0 : 1
      }
      this.customModalVisible = true
    },
    deleteCustomItem(item) {
      this.selectedData = item
      this.selectedDataType = 'custom'
      this.isShowWarning = true
    },
    closeCustomModal() {
      this.customModalVisible = false
    },
    formatCustomYaml() {
      const formatted = formatScrapeYaml(this.customForm.yaml_content)
      if (formatted === this.customForm.yaml_content) {
        this.$Message.info(this.$t('m_custom_scrape_format_unchanged'))
        return
      }
      this.customForm.yaml_content = formatted
      this.$Message.success(this.$t('m_custom_scrape_format_done'))
    },
    buildCustomPayload() {
      return {
        guid: this.customForm.guid,
        name: (this.customForm.name || '').trim(),
        yaml_content: this.customForm.yaml_content,
        enabled: this.customForm.enabled === 0 ? 0 : 1
      }
    },
    validateCustomYaml() {
      const payload = this.buildCustomPayload()
      if (!payload.name || !payload.yaml_content) {
        this.$Message.warning(this.$t('m_tips_required'))
        return
      }
      this.request('POST', this.apiCenter.customScrapeConfigCheck, payload, () => {
        this.$Message.success(this.$t('m_custom_scrape_validate_ok'))
      })
    },
    saveCustomItem() {
      const payload = this.buildCustomPayload()
      if (!payload.name || !payload.yaml_content) {
        this.$Message.warning(this.$t('m_tips_required'))
        return
      }
      const method = this.customForm.isAdd ? 'POST' : 'PUT'
      this.request(method, this.apiCenter.customScrapeConfig, payload, () => {
        this.customModalVisible = false
        this.getCustomScrapeList()
      })
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
.custom-scrape-yaml /deep/ textarea {
  font-family: Consolas, Monaco, "Courier New", monospace;
  font-size: 13px;
  line-height: 1.5;
}
</style>
