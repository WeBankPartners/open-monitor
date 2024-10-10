<template>
  <div class="page-component">
    <section class="page-header" >
      <div class="research c-dark" v-if="pageConfig.researchConfig">
        <research :pageConfig="pageConfig" :selectedData="selectedData" >
          <div slot="transmitExtraSearch">
            <slot name="extraSearch"></slot>
          </div>
          <div slot="transmitExtraBtn">
            <slot name="extraBtn"></slot>
          </div>
        </research>
      </div>
    </section>
    <section class="page-table c-dark">
      <tableTemp :table="pageConfig.table" :pageConfig="pageConfig" @sendIds="receiveIds" ref="refTest">
        <slot :name='pageConfig.table.isExtend.slot' v-if="pageConfig.table.isExtend"></slot>
      </tableTemp>
    </section>
    <div class="paging" v-if='pageConfig.pagination'>
      <template v-if="pageConfig.researchConfig">
        <pagination :pagination="pageConfig.pagination" :filters="pageConfig.researchConfig.filters" :pageUrl="pageConfig.CRUD"></pagination>
      </template>
      <!-- <template v-if="!pageConfig.researchConfig">
        <pagination :pagination="pageConfig.pagination"  :pageUrl="pageConfig.CRUD"></pagination>
      </template> -->
    </div>
  </div>
</template>
<script>
// import contentTitle from './title'
// import Notice from './notice'
import research from './research'
import tableTemp from './table'
import pagination from './pagination'

export default {
  name: 'page',
  data() {
    return {
      url: '',
      deleteSms: {
        title: '',
        show: false,
        btnText: '获取验证码',
        verifycode: '',
        isverify: false,
        btnDisabled: '',
        clock: '',
        phone: localStorage.phone,
        area_code: '',
        params: null,
        noSmscode: false
      }
    }
  },
  props: ['pageConfig', 'selectedData'],
  mounted() {
  },
  methods: {
    initData(url, params) {
      if (this.pageConfig.researchConfig){
        this.$root.$tableUtil.initTable(this, 'GET', url, params)
      } else {
        this.$parent.initData(this.pageConfig.CRUD, this.pageConfig)
      }
    },
    search() {
      // 搜索时清空已选中数据
      this.clearSelectedData()
      // 搜索时强制从第0条开始
      if (this.pageConfig.pagination){
        this.pageConfig.pagination.current = 1
      }
      this.$parent.initData(this.pageConfig.CRUD, this.pageConfig)
    },
    // 原始删除
    noSmsDelete(val, delTip){
      this.url = this.pageConfig.CRUD + '/' + val.id
      this.$deleting.deleteFunc({
        delTip,
        okCallback: this.ok
      })
    },
    // 原始确认删除
    ok() {
      this.$root.$tableUtil.deleteF(this, 'DELETE', this.url)
    },
    receiveIds(ids) {
      this.$parent.selectedData = ids
    },
    clearSelectedData() {
      this.$parent.selectedData = {checkedIds: []}
      this.$refs.refTest.initSelected()
    }
  },
  components: {
    research,
    // contentTitle,
    // Notice,
    tableTemp,
    pagination
  }
}
</script>

<style lang="less" scoped>
  .page-header {
    // margin-bottom: 20px;
    // padding-bottom: 12px;
    background-color: white;
    .research {
      padding-left: 10px;
    }
  }

  .page-table {
    background-color: white;
    // padding: 20px;
  }
  /*分页样式*/
  .paging {
    margin: 20px 20px 0 0;
    text-align: right;
  }

  /*删除模态框样式*/
  .delete-modal-wrap{
    .ivu-modal{
      width: 402px;
      height: 177px;
    }
    .ivu-btn-text{
      border: 1px solid #e9eaec;
    }
    .content{
      height: 89px;
      padding-top: 20px;
      padding-left: 22px;
      // display: flex;
      // justify-content: center;
      .warningIcon{
        width: 44px;
        height: 44px;
        border-radius: 50%;
        background-color: #fcac60;
        color: #ffffff;
        text-align: center;
        display: inline-block;
        vertical-align: middle;
        b{
          font-size: 28px;
          line-height: 44px;
        }
      }
      .baseline{
        display: inline-block;
        height: 100%;
        width: 1px;
        min-height: 50px;
        vertical-align: middle;
      }
      p{
        line-height: 1.4;
        padding-left: 22px;
        font-size: 16px;
        vertical-align: middle;
        display: inline-block;
        min-height: 22px;
        word-break: break-all;
        width: 290px;
      }
   }
  }
  //删除短信验证
   .Icon{
    display: inline-block;
    width: 36px;
    height: 36px;
    border-radius: 18px;
    background-color: #ffcb00;
    text-align: center;
    padding: 5px;
    div{
      .square{
        width: 4px;
        height: 16px;
        border-radius: 5px;
        background-color: #fff;
        margin-left: 11px;
        margin-top: 2px;
      }
      .circle{
        margin-top: 3px;
        width: 4px;
        height: 4px;
        background-color: #fff;
        border-radius: 2px;
        margin-left: 11px;
      }
    }
  }
  .Ititle{
    display: inline-block;
    margin-top: -10px;
    position: relative;
    top: -10px;
    left: 10px;
    width:260px;
    vertical-align: middle
  }
  .rebuilderrorcontent{
    margin-top:10px;
    p{
      color: #868686;
      font-size: 14px;
    }
  }
</style>
