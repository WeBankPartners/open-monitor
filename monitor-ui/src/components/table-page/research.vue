<template>
  <div class="research-component">
    <div class="research">
      <!-- <div v-if="pageConfig.researchConfig.superSearch" style="height: 32px">
        <byakugan :pageConfig="pageConfig" ref="superSearch"></byakugan>
        <div>
          <template v-for="(btn, index) in pageConfig.researchConfig.btn_group">
            <button :key='index' type="button" class="btn btn-sm"
                  @click="goToAction(btn.btn_func, pageConfig.researchConfig.filters)" :class="btn.class">
            <i :class="btn.btn_icon" ></i>
            {{btn.btn_name}}
          </button>
          </template>

        </div>
      </div> -->
      <form style="font-size:0;line-height:32px;" @submit.prevent>
        <template v-for="(input_condition, index_input_condition) in pageConfig.researchConfig.input_conditions" >
          <div :key="index_input_condition" class="research-div" >
            <input v-if="input_condition.type === 'input'" v-model.trim="pageConfig.researchConfig.filters[input_condition.value]"
                   :placeholder="$t(input_condition.placeholder)"
                   type="text"
                   @keyup.enter.stop="goToAction('search', pageConfig.researchConfig.filters, $event)"
                   class="form-control research-input c-dark"
                   @mouseover="activeInput(input_condition, index_input_condition)"
            />
            <i class="fa fa-plus-circle clearIcon" style="font-size:12px;"
               v-if="showClearIcon(input_condition, index_input_condition)" @click="clearInputCondition(input_condition)"
            ></i>
          </div>
        </template>
        <div class="research-div" style="font-size:12px;">
          <slot name="transmitExtraSearch"></slot>
        </div>
        <div class="button-div" style="font-size:12px;">
          <template v-for="(btn, index) in pageConfig.researchConfig.btn_group">
            <button :key='index' type="button" class="btn btn-sm"
                    @click="goToAction(btn.btn_func, pageConfig.researchConfig.filters, $event)" :class="btn.class"
            >
              <i :class="btn.btn_icon" ></i>
              {{$t(btn.btn_name)}}
            </button>
          </template>
          <div class="research-div" style="font-size:12px;">
            <slot name="transmitExtraBtn"></slot>
          </div>
        </div>
      </form>
      <!-- <div class="batch-operation" v-if="pageConfig.researchConfig.batch_btn_group"  @mouseleave="showBtnGroup=false">
        <div class="btn-group" role="group">
          <button type="button"  class="btn btn-sm btn-patch dropdown-toggle" @mouseover="showBtnGroup=!showBtnGroup" :disabled="isBatchBtnsActive">
            批量操作
          </button>
          <ul class="dropdown-btn" v-if="showBtnGroup">
            <template v-for="(btn, btnIndex) in pageConfig.researchConfig.batch_btn_group">
              <li class="filters-li" :key="btnIndex" @click="goToAction(btn.btn_func, btn.type)">{{btn.btn_name}}</li>
            </template>
          </ul>
        </div>
      </div> -->
    </div>
  </div>
</template>
<script>
// import byakugan from '../../components/common-temp/byakugan'

export default {
  name: 'research-component',
  data() {
    return {
      showBtnGroup: false,
      isBatchBtnsActive: true, //
      activeClearIcon: null, // 搜索区域当前关键字值
      activeClearIconNo: null, // 搜索区域序号(只包含简单搜索)
    }
  },
  props: ['pageConfig', 'selectedData'],
  watch: {
    selectedData(selectedData) {
      this.isBatchBtnsActive = (selectedData.checkedIds.length > 0 ? false : true)
    }
  },
  mounted() {

  },
  methods: {
    // 定位当前搜索框信息
    activeInput(input_condition, index_input_condition) {
      this.activeClearIcon = input_condition.value
      this.activeClearIconNo = index_input_condition
    },
    // 注销当前搜索框信息
    cancleActiveInput() {
      this.activeClearIcon = null
      this.activeClearIconNo = null
    },
    // 判断是否显示搜索框清除按钮
    showClearIcon(input_condition, index_input_condition) {
      const isInputEmpty = !this.$root.$validate.isEmpty_reset(this.pageConfig.researchConfig.filters[input_condition.value])
      return (index_input_condition === this.activeClearIconNo && isInputEmpty) ? true : false
    },
    // 搜索框清除按钮响应函数
    clearInputCondition(input_condition) {
      this.pageConfig.researchConfig.filters[input_condition.value] = ''
      this.cancleActiveInput()
      this.$parent.search()
    },
    goToAction(func, filters, event) {
      event.stopPropagation()
      if (func === 'search') {
        if (this.pageConfig.researchConfig.superSearch) {
          this.$refs.superSearch.search()
        } else {
          this.$parent[func](filters)
        }
      } else {
        this.$parent.$parent[func](filters)
      }
    }
  },
  components: {
    // byakugan
  }
}
</script>

<style lang="less" scoped>
  /*搜索组件整体样式*/
  .research {
    margin: 8px 0px;
    padding: 0px;
    text-align: left;
    padding-right: 150px;
    button {
      margin-bottom: 3px;
    }
  }
  /*控制每个搜索框样式*/
  .research-div{
    /*width: auto;*/
    // margin-bottom: 10px;
    display: inline-block;
    vertical-align: top;
  }

  .form-control {
    // 解决IE11 输入框无法与按钮同行问题
    display: inline-block;
    padding: 4px 18px 4px 7px;
  }
  /*取消选中样式*/
  .form-control:focus {
    box-shadow: none;
  }

  /*input框样式*/
  .research-input {
    width: 160px;
    height: 32px;
    font-size: 12px;
    margin-left: 8px;
    margin-right: 8px;
  }
  /*输入框清除按钮样式*/
  .clearIcon {
    position: relative;
    right: 28px;
    transform:rotate(45deg)
  }

  button {
    margin-left: 8px;
    margin-right: 8px;
  }
  button:focus {
    box-shadow: none;
  }
  .batch-operation {
    float: right;
    // margin-right: 100px;
    margin-top: -32px;
    .btn-group {
      margin-right: 5px;
      .btn-patch {
        background: @color-blue;
        border-radius: 4px;
        color: #ffffff;
        width: 109px;
      }
    }
  }
  .dropdown-btn {
    background: #FFFFFF;
    border: 1px solid @color-gray-E;
    border-radius: 4px;
    width: 109px;

    cursor: pointer;
    position: absolute;
    top: 100%;
    left: 0;
    z-index: 1000;
    float: left;
    list-style: none;
  }
  .btn{
    // width: 86px;
    height: 32px;
  }

  .filters-li {
    padding: 0 16px;
    line-height: 28px;
    text-align: left;
    cursor: pointer;
    color: #595959;
  }

  .filters-li:hover {
    background-color: @color-gray-F;
  }
  .button-div{
    display: inline-block;
  }
  // @media screen and (max-width: 1400px) {
  //   .button-div{
  //     display: block;
  //     margin-top: 8px;
  //     height: 32px;
  //   }
  // }
</style>
