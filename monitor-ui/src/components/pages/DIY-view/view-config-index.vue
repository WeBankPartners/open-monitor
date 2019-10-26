<template>
  <div class="">
    <div class="operational-zone">
      <Select v-model="model1" style="width:200px">
        <Option v-for="item in cityList" :value="item.value" :key="item.value">{{ item.label }}</Option>
      </Select>
      <button class="btn btn-sm btn-confirm-f" @click="addView">
        <i class="fa fa-plus"></i>{{$t('title.addViewTemplate')}}
      </button>
    </div>
    <section>
        <template v-for="(panalItem,panalIndex) in dataList">
          <div :key="panalIndex" class="panal-list">
            <Card>
              <p slot="title" class="panal-title">
                {{$t('title.templateName')}}:{{panalItem.name}}
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
    }
  },
  mounted(){
    this.viewList()
  },
  methods: {
    addPost () {
      this.JQ('#add_edit_Modal').modal('hide')
      let params = {
        name: this.modelConfig.addRow.name,
        cfg: ''
      }
      this.$httpRequestEntrance.httpRequestEntrance('POST','dashboard/custom/save', params, () => {
        this.viewList()
      })
    },
    addView () {
      this.modelConfig.isAdd = true
      this.JQ('#add_edit_Modal').modal('show')
    },
    removeTemplate (item) {
      let params = {id: item.id}
      this.$httpRequestEntrance.httpRequestEntrance('GET','dashboard/custom/delete', params, () => {
        this.$Message.success(this.$t('button.remove')+this.$t('tips.success'))
        this.viewList()
      })
    },
    viewList () {
      this.$httpRequestEntrance.httpRequestEntrance('GET','dashboard/custom/list', '', responseData => {
        this.dataList = responseData
      })
    },
    goToPanal(panalItem) {
      console.log(panalItem.cfg)
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
</style>
