<template>
  <div class="">
    <div class="operational-zone">
      <!-- <Select v-model="model1" style="width:200px">
        <Option v-for="item in cityList" :value="item.value" :key="item.value">{{ item.label }}</Option>
      </Select> -->
      <button class="btn btn-sm btn-confirm-f" @click="addView">
        <i class="fa fa-plus"></i>{{$t('button.addViewTemplate')}}
      </button>
      <button class="btn btn-sm btn-cancle-f" @click="setDashboard">
        {{$t('button.setDashboard')}}
      </button>
    </div>
    <section>
        <template v-for="(panalItem,panalIndex) in dataList">
          <div :key="panalIndex" class="panal-list">
            <Card>
              <p slot="title" class="panal-title">
                {{$t('title.templateName')}}:{{panalItem.name}}
                  <i class="fa fa-star" v-if="panalItem.main === 1" aria-hidden="true"></i>
              </p>
              <a slot="extra">
                <button class="btn btn-sm btn-confirm-f" @click="goToPanal(panalItem)">{{$t('button.view')}}</button>
                <button class="btn btn-sm btn-cancle-f" @click="removeTemplate(panalItem)">{{$t('button.remove')}}</button>
              </a>
              <ul class="panal-content">
                <li>
                  {{$t('title.updateTime')}}: {{panalItem.update_at}}
                </li>
              </ul>
            </Card>
          </div>
        </template>
      <!-- </ul> -->
    </section>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
    <ModalComponent :modelConfig="setDashboardModel">
      <div slot="setDashboard">  
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name lable-name-select">{{$t('title.templateName')}}:</label>
          <Select v-model="setDashboardModel.addRow.templateSelect" style="width:338px">
              <Option v-for="item in setDashboardModel.templateList" :value="item.value" :key="item.value">
              {{item.label}}</Option>
          </Select>
        </div>
      </div>
    </ModalComponent>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      cityList: [
        {
          value: 'New York',
          label: 'New York'
        },
        {
          value: 'Paris',
          label: 'Paris'
        }
      ],
      model1: '',
      dataList: [
      ],
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: 'title.viewTemplate',
        isAdd: true,
        config: [
          {label: 'tableKey.name', value: 'name', placeholder: 'tips.required', v_validate: 'required:true|min:2|max:60', disabled: false, type: 'text'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          name: null,
        },
      },
      setDashboardModel: {
        modalId: 'set_dashboard_modal',
        modalTitle: 'button.setDashboard',
        isAdd: true,
        saveFunc: 'setDashboardSave',
        config: [
          {name:'setDashboard',type:'slot'}
        ],
        addRow: {
          templateSelect: null
        },
        templateList: []
      }
    }
  },
  mounted(){
    this.viewList()
  },
  methods: {
    addPost () {
      this.$root.JQ('#add_edit_Modal').modal('hide')
      let params = {
        name: this.modelConfig.addRow.name,
        cfg: ''
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.template.save, params, () => {
        this.viewList()
      })
    },
    addView () {
      this.modelConfig.isAdd = true
      this.$root.JQ('#add_edit_Modal').modal('show')
    },
    removeTemplate (item) {
      let params = {id: item.id}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.template.delete, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.viewList()
      })
    },
    viewList () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.template.list, '', responseData => {
        this.setDashboardModel.templateList = []
        this.setDashboardModel.addRow.templateSelect = null
        this.dataList = responseData
        responseData.forEach((item) => {
          this.setDashboardModel.templateList.push({
            label: item.name,
            value: item.id
          })
          if (item.main === 1) {
            this.setDashboardModel.addRow.templateSelect = item.id
          }
        }) 
      })
    },
    setDashboard () {
      this.$root.JQ('#set_dashboard_modal').modal('show')
    },
    setDashboardSave () {
      let params = {id: this.setDashboardModel.addRow.templateSelect}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.template.templateSet, params, () => {
        this.$root.JQ('#set_dashboard_modal').modal('hide')
        this.$Message.success(this.$t('tips.success'))
        this.viewList()
      })
    },
    goToPanal(panalItem) {
      this.$router.push({name:'viewConfig',params:panalItem})
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
.operational-zone {
  margin: 16px;
}
.panal-list {
 margin: 16px;
}
.panal-title {
  color: @blue-2;
}
.panal-content {
  font-size: 12px;
}
.fa-star {
  color: @color-orange-F;
}
</style>
