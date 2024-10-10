<template>
  <div class="update-lists">
    <div class="modal-component">
      <div class="modal fade" id="custom_th_Modal" tabindex="-1" role="dialog" aria-labelledby="myModalLabel">
        <div class="modal-dialog" role="document">
          <div class="modal-content">
            <div class="modal-header">
              <h4 class="modal-title">自定义列表字段</h4>
            </div>
            <div class="modal-body">
              <form class="">
                <CheckboxGroup v-model="checkedItem">
                  <ul>
                    <template v-for="(item,index) in optionsTable">
                      <li :key="index" class="col-md-6 ul-li-style" style="float: left">
                        <label class="ivu-checkbox-wrapper ivu-checkbox-group-item ivu-checkbox-wrapper-checked">
                          <!--固定显示-开始-->
                          <template v-if="item.frozen">
                            <span class="ivu-checkbox ivu-checkbox-checked">
                              <span class="checkbox-checked-disabled"></span>
                            </span>
                            <span>{{item.title}}</span>
                          </template>
                          <!--固定显示-结束-->
                          <!--任意显示-开始-->
                          <template v-else>
                            <span :class="{'ivu-checkbox': true ,'ivu-checkbox-checked': item.display}"
                                  @click="item.display = !item.display"
                            >
                              <span class="ivu-checkbox-inner"></span>
                              <input type="checkbox" class="ivu-checkbox-input" :value="index"/>
                            </span>
                            <span>{{item.title}}</span>
                          </template>
                        <!--任意显示-结束-->
                        </label>
                      </li>
                    </template>
                  </ul>
                </CheckboxGroup>
              </form>
            </div>
            <div class="modal-footer">
              <Button data-dismiss="modal">{{ $t('m_button_cancel') }}</Button>
              <Button type="primary" @click="changeTableTr">{{ $t('m_save') }}</Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  name: 'update-lists',
  data() {
    return {
      checkedItem: [], // 选中的列
      optionsTable: [],
    }
  },
  props: ['table'],
  mounted() {

  },
  methods: {
    showCustomTh() {
      this.optionsTable = []
      this.table.tableEle.forEach((item, index) => {
        if (item.display === true) {
          this.checkedItem.push(index)
        }
        const obj = {}
        for (const key in item){
          obj[key] = item[key]
        }
        this.optionsTable.push(obj)
      })
      this.$root.JQ('#custom_th_Modal').modal('show')
    },
    changeTableTr() {
      this.table.tableEle = this.optionsTable
      // 将自定义列表信息更新至vuex缓存
      const column = {
        name: this.$router.history.current.name,
        params: this.table.tableEle,
      }
      this.$parent.changeTdNumber()
      // this.tdNumber = this.gettdsLength()
      this.$root.$store.commit('catchColumn', column)
      this.$root.JQ('#custom_th_Modal').modal('hide')
    },
  },
  components: {}
}
</script>

<style lang="less" scoped>
  .modal-dialog {
    top: 25%;
  }
  .modal-body {
    padding: 16px 8px;
    form {
      margin: 0 30px;
    }
  }
  .modal-header {
    height: 55px;
    align-items: center;
    .modal-title{
      font-size: 16px;
      color: rgba(0,0,0,0.85);
    }
  }

  .ul-li-style{
    float: left;
  }

  .btn {
    width: 65px;
    height: 32px;
    padding: 4px 0;
  }

  button:focus {
    box-shadow: none;
  }

  #custom_th_Modal ul li{
    list-style: none;
    .ivu-checkbox-checked{
      .ivu-checkbox-inner {
        border-color: @color-blue !important;
        background-color: @color-blue !important;
      }
    }
    .checkbox-checked-disabled {
      border-color: @color-gray-E !important;
      background-color: @color-gray-E !important;
      display: inline-block;
      width: 14px;
      height: 14px;
      position: relative;
      top: 0;
      left: 0;
      border: 1px solid #dddee1;
      border-radius: 2px;
      background-color: #fff;
    }

    .checkbox-checked-disabled:after {
    content: '';
    display: table;
    width: 4px;
    height: 8px;
    position: absolute;
    top: 1px;
    left: 4px;
    border: 2px solid #fff;
    border-top: 0;
    border-left: 0;
    transform: rotate(45deg) scale(1);
  }
  }
</style>
