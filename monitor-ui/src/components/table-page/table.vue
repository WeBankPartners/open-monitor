<template>
  <div class="table-template">
    <div class="table-operation">
      <div class="btn-group-logo">

        <!-- <Tooltip placement="top" v-if="pageConfig.titleConfig || pageConfig.noticeConfig || pageConfig.researchConfig">
          <img src="../../assets/table/refresh.svg" alt="列表刷新入口"  @click="refreshTable" class="table-operation-icon">
          <div slot="content">
            刷新
          </div>
        </Tooltip>

        <Tooltip placement="top" v-if="pageConfig.titleConfig || pageConfig.noticeConfig || pageConfig.researchConfig">
          <img src="../../assets/table/listEdit.svg" alt="列管理入口"  @click="showCustomTh" class="table-operation-icon">
          <div slot="content">
            自定义列表字段
          </div>
        </Tooltip> -->
      </div>
    </div>
    <div class="ui-table-container">
      <table class="table table-hover" style="table-layout:auto;">
        <!-- table-striped -->
        <thead>
          <tr>
            <th v-if="table.selection" class="th-border-bottom c-dark">
              <a href="javascript:void 0" @click="allSelect()">
                <span class="item-check-btn" :class="{'check': allSelectBtn}">
                  <svg class="icon icon-ok"><use xlink:href="#icon-ok"/></svg>
                </span>
              </a>
            </th>
            <template v-for="(item,tableEleIndex) in table.tableEle">
              <th style='min-width:100px;' :style="item.style" class="th-border-bottom c-dark" v-if="item.display" :key='tableEleIndex' >
                <span v-if="!item.copyable">{{$t(item.title)}}</span>
                <span v-if="item.copyable" :style="item.copyable ? 'cursor: pointer;position: relative;' : ''">
                  <Tooltip placement="bottom" :delay="500">
                    {{$t(item.title)}}
                    <div slot="content" class="tooltip-clipboard" @click="doCopy(tableEleIndex,item)">
                      <Icon type="clipboard"></Icon>
                      <span>
                        复制列
                      </span>
                    </div>
                  </Tooltip>
                </span>
                <span class="ivu-table-sort" v-if="item.sortable">
                  <i class="ivu-icon ivu-icon-arrow-up-b"  :class="{on: getColumn(tableEleIndex)._sortType === 'asc'}" @click="sort(item.value, tableEleIndex, '+')"></i>
                  <i class="ivu-icon ivu-icon-arrow-down-b" :class="{on: getColumn(tableEleIndex)._sortType === 'desc'}" @click="sort(item.value, tableEleIndex, '-')"></i>
                </span>
                <span class="" v-if="item.filterable && !item.filterable.multifilter">
                  <Poptip content="" placement="bottom" :transfer="tip_transfer" v-model="visibleFilter">
                    <i class="ivu-icon ivu-icon-funnel" @click="filterIcon(tableEleIndex)" :class="{on: getColumnFilter(tableEleIndex)._isFilter === 'Y'}"></i>
                    <div slot="content" v-if="pageConfig.table.tableEle[tableEleIndex]._isShowFilterPop" class='popfilter'>
                      <div class="" style="padding: 10px 16px 4px;">
                        <input type="text" v-model="filterParam" @click.stop.prevent=""
                               @keyup="searchFiltersContent" placeholder="请搜索"
                               @keyup.8="searchFiltersContent" class="tip-search"
                        />
                      </div>
                      <ul>
                        <li class="filters-li" @click="filterTable(item, tableEleIndex, 'All')">全部</li>
                        <!--<template v-for="(filter,j) in item.filterable.filterLists">-->
                        <template v-for="(filter,j) in filterListsBySearch">
                          <li class="filters-li" @click="filterTable(item, tableEleIndex, filter)" :key='j'
                              :class="{active: getColumnFilterActiveno(tableEleIndex)._activeKey === filter.value}"
                          >{{filter.label}}</li>
                        </template>
                      </ul>
                    </div>
                  </Poptip>
                </span>
                <span class="" v-if="item.filterable && item.filterable.multifilter">
                  <Poptip content="" placement="bottom" :transfer="tip_transfer" v-model="multiVisibleFilter">
                    <i class="ivu-icon ivu-icon-funnel" @click="filterIcon(tableEleIndex)" :class="{on: getColumnFilter(tableEleIndex)._isFilter === 'Y'}"></i>
                    <div slot="content" v-show="pageConfig.table.tableEle[tableEleIndex]._isShowFilterPop" class='multiFilter'>
                      <!-- <div class="" style="padding: 10px 16px 4px;">
                      <input type="text" v-model="filterParam" @click.stop.prevent=""
                            @keyup="searchFiltersContent" placeholder="请搜索"
                            @keyup.8="searchFiltersContent" class="tip-search">
                    </div> -->
                      <div class="popfilter">
                        <CheckboxGroup v-model="pageConfig.table.tableEle[tableEleIndex].multiFilterParams">
                          <template v-for="(filter,j) in filterListsBySearch">
                            <Checkbox :label="filter.value" :key="j">{{filter.label | interceptParams(60)}}</Checkbox>
                          </template>
                        </CheckboxGroup>
                      </div>
                      <div class="tooltip-footer">
                        <label class="tooltip-footer-btn" @click="multiFilter(item, tableEleIndex)">筛选</label>
                        <label class="tooltip-footer-btn" @click="resetMultiFilter(item, tableEleIndex)">重置</label>
                      </div>
                    </div>
                  </Poptip>
                </span>
              </th>
            </template>

            <th style="width: 140px;" class="th-border-bottom c-dark" v-if="table.btn.length !== 0"><div style="width:130px">{{$t('m_table_action')}}</div></th>
            <th style="width: 151px;" class="th-border-bottom c-dark handleSty" v-if="table.btn.length !== 0 && table.handleFloat">{{$t('m_table_action')}}</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="(value,tableDataIndex) in table.tableData">
            <tr :key="tableDataIndex">
              <!--列表是否显示勾选框--开始-->
              <td class="td-center" v-if="table.selection">
                <div class="cart-item-check">
                  <a href="javascript:void 0" class="item-check-btn" :class="{'check': value.checked}"
                     @click="selectProduct(value)"
                  >
                    <svg class="icon icon-ok">
                      <use xlink:href="#icon-ok"/>
                    </svg>
                  </a>
                </div>
              </td>
              <!--列表是否显示勾选框--结束-->

              <template  v-for="(val,i) in table.tableEle">
                <td class="c-dark" v-if="val.display" :key="i" :class="ellipsis(value, val) ? 'tdoverflow' : ''"><!--ellipsis是否显示省略-->
                  <div class="extendStyle" v-if='table.isExtend && i === firstShow'>
                    <a v-show="tableDataIndex !== currentActive" @click="loadDetail(value,tableDataIndex)"><i class="ivu-icon ivu-icon-ios-arrow-forward" style="font-size: 20px;"></i></a>
                    <a v-show="tableDataIndex === currentActive" @click="loadDetail(value,tableDataIndex)" class="active"><i class="ivu-icon ivu-icon-ios-arrow-forward" style="font-size: 20px;"></i></a>
                  </div>
                  <!-- 显示自定义tip -->
                  <Tooltip v-if="val.toolTips && renderValue(value,val.toolTips)" :content="value[val.toolTips]" class="cell" :delay=1000  placement='bottom-start'>
                    <a @click="shadow(value, val)" v-if="'shadow' in val" class="extend-shadow poplableSty" :style="val.toolTips.style">{{renderValue(value, val)}}</a>
                    <span v-if="!('shadow' in val)" :style="val.toolTips.style" class="poplableSty">{{renderValue(value, val)}}</span>
                    <div slot="content">
                      <!--<div class="Newline">{{renderValue(value, val)}}</div>-->
                      <div class="Newline" v-if="val.toolTips.type === 'text'">{{renderValue(value,val.toolTips)}}</div>
                      <div class="Newline" v-if="val.toolTips.type === 'json'">
                        <pre style="color:white;overflow-y:auto;max-height:300px">{{renderValue(value,val.toolTips)}}</pre>
                      </div>
                    </div>
                  </Tooltip>
                  <!-- 显示省略且未自己配置tip情况下，自动添加tip,tip内容为整体内容 -->
                  <Tooltip v-if="ellipsis(value, val) && !val.toolTips && !val.renderContent" :content="renderValue(value, val)" class="cell" :delay=1000  placement='bottom-start'>
                    <a @click="shadow(value, val)" v-if="'shadow' in val" class="extend-shadow poplableSty">{{renderValue(value, val)}}</a>
                    <span v-if="!('shadow' in val)" class="poplableSty">{{renderValue(value, val)}}</span>
                    <div slot="content">
                      <div class="Newline">{{renderValue(value,val)}}</div>
                    </div>
                  </Tooltip>
                  <!-- 不显示省略情况下 -->
                  <div class="cell" v-if="!ellipsis(value, val) && (!val.toolTips || (val.toolTips && !renderValue(value,val.toolTips)))"
                       @mouseover="isShowEdit(tableDataIndex,i)"
                  >

                    <!--列表内容有跳转-开始-->
                    <a v-if="'shadow' in val" @click="shadow(value, val)"  class="extend-shadow">{{renderValue(value, val)}}</a>
                    <!--列表内容有跳转-结束-->

                    <!--列表内容无跳转-开始-->
                    <template v-if="!('shadow' in val) && !(val.tags)">
                      <span v-if="!(isShowEditXX(val,tableDataIndex, i) && isShowEditInput)">{{renderValue(value, val)}}</span>
                      <input type="text" v-else id="editInput" v-model="editContent" class="input-style"
                             @keyup.enter="inputBlur(val,value)" @blur.prevent="inputBlur(val,value)"
                      />
                      <!-- <i v-if="isShowEditXX(val,tableDataIndex, i)&&!isShowEditInput" @click="showEditInput(value,val,tableDataIndex, i)"
                      class="fa fa-pencil cell-edit" aria-hidden="true"></i> -->
                      <i v-if="val.editable" @click="showEditInput(value,val,tableDataIndex, i)"
                         class="fa fa-pencil cell-edit" aria-hidden="true"
                      ></i>

                    </template>
                    <!--列表内容无跳转-结束-->
                    <template v-if="val.hasIcon && ( value[val.hasIcon.judgeflag] ^ val.hasIcon.reverse)">
                      <Tooltip placement="bottom">
                        <i class="fa cell-icon" :class="val.hasIcon.icon"
                           @click="goToOpe(val.hasIcon,value)" aria-hidden="true"
                           :title="val.hasIcon.title"
                        ></i>
                        <div slot="content">
                          {{computeTooltipContent(value, val)}}
                        </div>
                      </Tooltip>
                    </template>

                    <!--列表标签适配--开始-->
                    <template v-if="val.tags">
                      <div style="display: flex;flex-direction: row;flex-wrap: wrap;" :style="val.tags.style">
                        <template v-for="(tagsData, tagsIndex) in renderValue(value, val)">
                          <span class="tag-f" :key="tagsIndex">{{tagsData.label}}</span>
                        </template>
                      </div>
                    </template>

                    <!--列表标签适配--结束-->
                    <!-- 列表内容为slot -->
                    <span v-if="val.slot">
                      <slot :name="val.slot" :node="value"></slot>
                    </span>

                  </div>
                  <template v-if="val.renderContent">
                    <div class="render-content" v-html="renderValue(value, val)"></div>
                  </template>
                </td>
              </template>

              <!--操作区--开始-->
              <!-- <td class="td-center td-operation c-dark" v-if="table.btn.length != 0"> -->
              <!-- <div style="width:140px"> -->
              <!-- <template v-for="(btn_val,btn_i) in table.btn">
                <span :id="btn_val.btn_func" :key="btn_i"
                      v-if="btn_val.btn_name !='more' && !btn_val.render"
                      class="btn-operation"
                      @click="goToOpe(btn_val,value)">
                  {{btn_val.btn_name}}
                </span>
                <span :id="btn_val.btn_func" :key="btn_i + '2'"
                      v-if="btn_val.btn_name !='more' && btn_val.render"
                      class="btn-operation"
                      @click="goToOpe(btn_val,value,tableDataIndex)">
                      {{renderValue(value, btn_val)}}
                </span>
                <span class="btn-pipe" v-if="btn_i != table.btn.length - 1">
                  |
                </span>
                <Poptip content="" placement="bottom" trigger="hover">
                  <div class="batch-operation" v-if="btn_val.btn_name ==='more'">
                    <div class="btn-group" role="group" @mouseover="btnMore(tableDataIndex)">
                    <span class="btn-operation">
                      更多
                      <Icon type="chevron-down"></Icon>
                    </span>
                    </div>
                  </div>
                  <div slot="content" class="ui-ul-list" v-if="isShowMoreBtnTip">
                    <ul v-if="showMoreNumber === tableDataIndex">
                      <template v-for="(btn, moreIndex) in filterMoreBtn(value)">
                        <li class="filters-li" :key="moreIndex" v-if="btn.isShowBtn" @click="getToMoreAction(btn.btn_func,value,tableDataIndex)">{{btn.btn_name}}</li>
                      </template>
                      <li v-if="isNoActions" style="color: red" @click="noActions">无可用操作</li>
                    </ul>
                  </div>
                </Poptip>
              </template> -->
              <!-- </div> -->
              <!-- </td> -->
              <td class="td-center td-operation handleSty c-dark" v-if="table.btn.length !== 0 && table.handleFloat" style="padding:9px 0px">
                <div style="width: 151px;padding-left: 8px;height: 21px;">
                  <template v-if="operationsFormat(value, tableDataIndex)">
                  </template>
                  <template v-if="table.tableData[tableDataIndex]._operatons.length <= 2">
                    <template v-for="(btn_val,btn_i) in table.tableData[tableDataIndex]._operatons">
                      <span :id="btn_val.btn_func" :key="btn_i"
                            v-if="!btn_val.render"
                            class="btn-operation"
                            :style="{color: btn_val.color}"
                            @click="goToOpe(btn_val,value)"
                      >
                        {{$t(btn_val.btn_name)}}
                      </span>
                      <span :id="btn_val.btn_func" :key="btn_i + '2'"
                            v-if="btn_val.render"
                            class="btn-operation"
                            :style="{color: btn_val.color}"
                            @click="goToOpe(btn_val,value)"
                      >
                        {{renderValue(value, btn_val)}}
                      </span>
                      <span class="btn-pipe" :key="btn_i + '1'" v-if="btn_i !== 1 & table.tableData[tableDataIndex]._operatons.length !== 1">
                        |
                      </span>
                    </template>
                  </template>
                  <template v-else>
                    <span :id="table.tableData[tableDataIndex]._operatons[0].btn_func"
                          v-if="!table.tableData[tableDataIndex]._operatons[0].render"
                          class="btn-operation"
                          @click="goToOpe(table.tableData[tableDataIndex]._operatons[0],value)"
                    >
                      {{$t(table.tableData[tableDataIndex]._operatons[0].btn_name)}}
                    </span>
                    <span :id="table.tableData[tableDataIndex]._operatons[0].btn_func" :key="btn_i + '2'"
                          v-if="table.tableData[tableDataIndex]._operatons[0].render"
                          class="btn-operation"
                          @click="goToOpe(table.tableData[tableDataIndex]._operatons[0],value)"
                    >
                      {{renderValue(value, table.tableData[tableDataIndex]._operatons[0])}}
                    </span>
                    <span class="btn-pipe">
                      |
                    </span>
                    <Poptip content="" placement="bottom" trigger="hover">
                      <div class="batch-operation">
                        <div class="btn-group" role="group" @mouseover="btnMore(tableDataIndex)">
                          <span class="btn-operation">
                            {{$t('m_table_more')}}
                            <Icon type="chevron-down"></Icon>
                          </span>
                        </div>
                      </div>
                      <div slot="content" class="ui-ul-list" v-if="isShowMoreBtnTip">
                        <ul v-if="showMoreNumber === tableDataIndex">
                          <template v-for="(btn, moreIndex) in table.tableData[tableDataIndex]._operatons">
                            <li class="filters-li" :key="moreIndex" v-if="moreIndex !== 0" @click="goToOpe(btn,value)" :style="{color: btn.color}">{{$t(btn.btn_name)}}</li>
                          </template>
                          <li v-if="isNoActions" style="color: red" @click="noActions">无可用操作</li>
                        </ul>
                      </div>
                    </Poptip>
                  </template>
                  <div>

                  </div>
                </div>
              </td>
            <!--操作区--结束-->
            </tr>
            <tr v-if="tableDataIndex === currentActive" :key="tableDataIndex + 0.5" class='bgc'>
              <td :colspan='tdNumber' id='extend1' class="c-dark-gray">
                <tdSlot :name='table.isExtend.slot'> <slot :name='table.isExtend.slot'></slot> </tdSlot>
              <!-- tdslot在main.js中定义 -->
              </td>
            </tr>
          </template>
          <tr v-if="table.tableData.length < 1" class='nodataSty c-dark'>
            <td :colspan='tdNumber' v-if='!table.noDataTip'>{{$t('m_table_noDataTip')}}</td>
            <td :colspan='tdNumber' v-if='table.noDataTip'>
              <template v-if='typeof table.noDataTip === "string"'>
                <span>{{$t('m_table_noDataTip')}},您可以</span><span v-html="table.noDataTip">{{table.noDataTip}}</span>
              </template>
              <template v-if='typeof table.noDataTip === "object"'>
                <span>{{$t('m_table_noDataTip')}},您可以</span><a @click="goToConfig(table.noDataTip)">{{table.noDataTip.linkword}}</a>{{table.noDataTip.name}}
              </template>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <!--自定义列模态框-->
    <!-- <updateLists :table ="table" ref="updateList"></updateLists> -->
  </div>
</template>
<script>
// import updateLists from './table/update-lists'

export default {
  data() {
    return {
      tip_transfer: true, // 作用参见iview tip transfer属性
      isShowMoreBtnTip: false, // 主动控制更多按钮tip是否显示
      showMoreNumber: null, // 控制更多按钮显示序号
      allSelectBtn: false,
      allSelectflagActivate: true, // 全选区域是否生效标志位
      selectedData: [],
      visibleFilter: false, // 控制筛选poptips是否显示
      filterParam: '', // tip搜索内容
      filterLists: [], // 待过滤所有选项
      filterListsBySearch: [], // 已过滤所有选项
      clipboardMsg: '', // 缓存复制至剪切板前内容
      currentActive: -1,// 当前列表展开index
      tdNumber: 0,
      firstShow: 0,

      moreBtnGroup: [],

      isNoActions: false,
      showEditRowNumber: null, // 当前应显示编辑按钮的行号
      showEditColumnNumber: null, // 当前应显示编辑按钮的列号
      isShowEditInput: false, // 控制是否显示编辑框
      editContent: null, // 待编辑内容
      activeEditInputNo: [], // 编辑状态的行/列号

      // 多选配置-开始
      multiVisibleFilter: false, // 控制筛选poptips是否显示
      // 多选配置-结束
    }
  },
  props: {
    table: {
      type: Object,
      required: true
    },
    pageConfig: {
      type: Object,
      required: true
    }
  },
  watch: {
    // 因页面keep-alive关系，在过滤框active状态下会出问题，此处在发生路由跳转时，
    // 强制隐藏过滤框
    $route(){
      this.visibleFilter = false
      this.multiVisibleFilter = false
    }
  },
  // 清除列表选中项，易诱发bug
  activated() {
    this.initSelected()
  },
  mounted() {
    // this.$root.JQ('[data-toggle="tooltip"]').tooltip()
  },
  created() {
    // 判断vuex是否包含自定义列表信息，获取相关信息赋给table
    // if (this.$root.$validate.isEmpty(this.$root.$store.state[this.$router.history.current.name])) {
    //   this.table.tableEle = this.$root.$store.state[this.$router.history.current.name]
    // }
    this.tdNumber = this.gettdsLength()
  },
  updated(){
    this.tdNumber = this.gettdsLength()
    this.currentActive = this.$root.$store.state.tableExtendActive
    this.$root.JQ('.ivu-tooltip-popper').css('display','none') // 详情跳转回table页面,禁止tooltip显示完整字段
  },
  methods: {
    // 依据数据处理可用操作，需优化
    operationsFormat(value, tableDataIndex) {
      if (this.$root.$validate.isEmpty_reset(this.table.filterMoreBtn)) {
        this.table.tableData[tableDataIndex]._operatons = this.table.btn
      } else {
        const moreBtnGroup_show = this.$parent.$parent.filterMoreBtn(value)
        const resBtns = []
        for (const btn of this.table.btn) {
          if (moreBtnGroup_show.includes(btn.btn_func)) {
            resBtns.push(btn)
          }
        }
        this.table.tableData[tableDataIndex]._operatons = resBtns
      }
    },
    filterMoreBtn(value) {
      // 初次进入获取更多按钮组信息
      if (!this.$root.$validate.isEmpty(this.moreBtnGroup)) {
        for (const moreBtn of this.table.btn) {
          if (moreBtn.btn_name === 'more') {
            this.moreBtnGroup = moreBtn.more
          }
        }
      }
      // 按钮组isShowBtn属性置为true
      for (const index in this.moreBtnGroup) {
        this.moreBtnGroup[index].isShowBtn = true
      }
      // 判断是否需控制按钮时
      if (this.$root.$validate.isEmpty(this.table.filterMoreBtn)) {
        // 按钮组isShowBtn属性置为true
        for (const index in this.moreBtnGroup) {
          this.moreBtnGroup[index].isShowBtn = false
        }
        // 获取需要显示的按钮
        const moreBtnGroup_show = this.$parent.$parent.filterMoreBtn(value)
        if (moreBtnGroup_show.length > 0) {
          this.isNoActions = false
        } else {
          this.isNoActions = true
        }
        for (const show_btn of moreBtnGroup_show) {
          for (const index in this.moreBtnGroup) {
            if (this.moreBtnGroup[index].btn_func === show_btn) {
              this.moreBtnGroup[index].isShowBtn = true
            }
          }
        }
      }
      return this.moreBtnGroup
    },
    noActions() {
      this.$Message.warning('该数据无可用操作！')
    },
    // TODO 计算icon Tooltips中内容，直接取值会报错
    computeTooltipContent(value, val) {
      if (this.$root.$validate.isEmpty(val.hasIcon)) {
        if (val.hasIcon.tip.includes('.')) {
          const arr = val.hasIcon.tip.split('.')
          let res = value
          for (const item of arr) {
            res = res[item]
          }
          return res
        }

        return value[val.hasIcon.tip]

      }
    },
    // 排序响应函数
    sort(key , i, sort) {
      const orders = sort + key
      this.pageConfig.pagination['__orders'] = orders
      this.pageConfig.pagination.current = 1
      this.$parent.$parent.initData(this.pageConfig.CRUD, this.pageConfig)
      for (const ele in this.pageConfig.table.tableEle) {
        this.pageConfig.table.tableEle[parseInt(ele)]._sortType = 'normal'
      }
      this.pageConfig.table.tableEle[i]._sortType = sort ==='-'? 'desc': 'asc'
    },
    // 获取排序状态active状态
    getColumn(index) {
      return {_sortType: this.pageConfig.table.tableEle[index]._sortType}
    },
    filterIcon(columnNo) {
      // 清空tip中搜索条件
      this.filterParam = ''
      if (this.pageConfig.table.tableEle[columnNo].filterable.remoteFilters) {
        this.getRemoteFilterOptions(columnNo)
      } else {
        this.filterLists = this.pageConfig.table.tableEle[columnNo].filterable.filterLists
        this.filterListsBySearch = this.pageConfig.table.tableEle[columnNo].filterable.filterLists
      }
      for (const item in this.pageConfig.table.tableEle) {
        this.pageConfig.table.tableEle[item]._isShowFilterPop = false
      }
      this.pageConfig.table.tableEle[columnNo]._isShowFilterPop = true
    },
    // 远程获取过滤条件
    getRemoteFilterOptions(columnNo) {
      console.error(columnNo)
      // this.filterLists = []
      // let filterAPI_string = this.pageConfig.table.tableEle[columnNo].filterable.filterAPI
      // let filterAPI = this.$root.$validate.valueFromExpression(this.$root.apiCenter, filterAPI_string)
      // let params = ''
      // if (!this.$root.$validate.isEmpty_reset(this.pageConfig.table.tableEle[columnNo].filterable.params)) {
      //   params = this.pageConfig.table.tableEle[columnNo].filterable.params
      // }
      // this.$root.$httpRequestEntrance.httpRequestEntrance('GET', filterAPI, params , (responseData) => {
      //   for (let item of responseData.data) {
      //     let label = ''
      //     if (!this.$root.$validate.isEmpty_reset(this.pageConfig.table.tableEle[columnNo].filterable.displayFormat)) {
      //       let str = this.pageConfig.table.tableEle[columnNo].filterable.displayFormat
      //       let reg = /\$\{[^(\$\{)]*\}/g
      //       let arr = str.match(reg).map(it => {
      //         return it.split('').slice(2,-1).join('')
      //       })
      //       let i = -1
      //       label = str.replace(reg,'`').split('').map((it) => {
      //         if (it === '`') {
      //           i++
      //           return this.$root.$validate.valueFromExpression(item, arr[i])
      //         } else {
      //           return it
      //         }
      //       }).join('')
      //     } else {
      //       label = item[this.pageConfig.table.tableEle[columnNo].filterable.displayName]
      //     }
      //     if (this.$root.$validate.isEmpty(label)) {
      //       this.filterLists.push({label: label,
      //         value: item[this.pageConfig.table.tableEle[columnNo].filterable.filterValue]})
      //     }
      //   }
      //   this.filterListsBySearch = this.filterLists
      // }, false)
    },
    // 列筛选功能
    filterTable(ele, i, valx) {
      if (valx === 'All') {
        this.pageConfig.table.tableEle[i]._isFilter = 'N'
        delete this.pageConfig.pagination[ele.filterable.filterParam]
        if (Object.prototype.hasOwnProperty.call(this.pageConfig.pagination, ele.value)) {
          delete this.pageConfig.pagination[ele.value]

        }
      } else {
        this.pageConfig.table.tableEle[i]._isFilter = 'Y'
        this.pageConfig.pagination[ele.filterable.filterParam] = valx.value
      }
      this.visibleFilter = false
      // 设置当前状态数据key
      this.pageConfig.table.tableEle[i]._activeKey = valx.value

      this.pageConfig.pagination.current = 1
      this.$parent.$parent.initData(this.pageConfig.CRUD, this.pageConfig)
    },
    // 多选筛选控制
    multiFilter(ele, i) {
      this.pageConfig.table.tableEle[i]._isFilter = 'Y'
      this.pageConfig.pagination[ele.filterable.filterParam] = this.pageConfig.table.tableEle[i].multiFilterParams
      this.multiVisibleFilter = false

      this.pageConfig.pagination.current = 1
      this.$parent.$parent.initData(this.pageConfig.CRUD, this.pageConfig)

      // .multiXX
    },
    // 多选重置参数
    resetMultiFilter(ele, i) {
      this.pageConfig.table.tableEle[i]._isFilter = 'N'
      this.pageConfig.table.tableEle[i].multiFilterParams = []
      this.pageConfig.pagination[ele.filterable.filterParam] = this.pageConfig.table.tableEle[i].multiFilterParams
      this.multiVisibleFilter = false
      this.pageConfig.pagination.current = 1
      this.$parent.$parent.initData(this.pageConfig.CRUD, this.pageConfig)
    },
    // 获取排序状态active状态
    getColumnFilter(index) {
      return {_isFilter: this.pageConfig.table.tableEle[index]._isFilter}
    },
    // 获取排序状态active状态中的当前状态
    getColumnFilterActiveno(index) {
      return {_activeKey: this.pageConfig.table.tableEle[index]._activeKey}
    },
    initSelected() {
      this.selectedData = []
      this.allSelectBtn = false
      this.allSelectflagActivate = true
      this.$emit('sendIds', {
        checkedIds: []
      })
    },
    /* ******** 普通按钮 ***** 第一个参数是id ******** */
    goToOpe(btn_val, value) {
      this.$parent.$parent[btn_val.btn_func](value)
    },
    // // 更多选项中按钮响应 ******** 第一个参数是对象 ********
    // getToMoreAction (btn_val, value, index) {
    //   // 点击后隐藏poptip
    //   this.isShowMoreBtnTip = false
    //   this.$parent.$parent[btn_val.btn_func](value[this.table.primaryKey], index)
    // },
    // 控制显示当前more按钮选项
    btnMore(index) {
      // 悬停显示poptip
      this.isShowMoreBtnTip = true
      this.showMoreNumber = index
    },
    selectProduct(item) {
      this.confirmSelected(item)
      if (typeof item.checked === 'undefined') {
        this.$set(item, 'checked', true)
      } else {
        item.checked = !item.checked
      }
      // 全部选中则全选点亮，相反
      let checkAllFlags = true
      this.table.tableData.forEach(function (value) {
        checkAllFlags = checkAllFlags && value.checked
      })
      this.allSelectBtn = checkAllFlags
    },
    // 点击全选
    allSelect() {
      this.selectedData = []
      if (this.allSelectflagActivate) {
        this.table.tableData.forEach(item => {
          this.confirmSelected(item)
        })
      } else {
        this.selectedData = []
      }
      this.allSelectBtn = this.allSelectflagActivate
      this.allSelectflagActivate = !this.allSelectflagActivate
      this.table.tableData.forEach(item => {
        if (typeof item.checked === 'undefined') {
          this.$set(item, 'checked', this.allSelectBtn)
        } else {
          item.checked = this.allSelectBtn
        }
      })
      this.$emit('sendIds', {
        checkedIds: this.selectedData
      })
    },
    removeByValue(arr, val) {
      for (let i = 0; i < arr.length; i++) {
        if (arr[i] === val) {
          arr.splice(i, 1)
          break
        }
      }
    },
    confirmSelected(item) {
      const id = item[this.table.primaryKey]
      if (this.$root.JQ.inArray(id, this.selectedData) === -1) {
        this.selectedData.push(id)
      } else {
        this.removeByValue(this.selectedData, id)
      }
      this.$emit('sendIds', {
        checkedIds: this.selectedData
      })
    },
    renderValue(item, config) {
      const func=config.render
      if (func){
        return func(item)
      }
      const expr=config.value || config.val
      let item_tmp = item
      const attrs=expr.split('.')
      let n=0
      if (!this.$root.$validate.isEmpty(item)){
        return ''
      }
      for (n in attrs){
        if (this.$root.$validate.isEmpty(item_tmp) && attrs[n] in item_tmp){
          item_tmp = item_tmp[attrs[n]]
        } else {
          return ''
        }
      }
      return item_tmp
    },
    // table内跳转
    shadow(item, val) {
      this.currentActive = -1
      this.$root.$store.commit('changeTableExtendActive',this.currentActive)
      const id = this.$root.$validate.valueFromExpression(item, val.shadow.key)
      const router = {
        name: val.shadow.path,
        params: {
          item,
          pageConfig: this.pageConfig,
          id
        }
      }
      if (val.shadow.query){// 根据key值组装传参对象
        router.query = val.shadow.query.map(key => ({[key]: this.$root.$validate.valueFromExpression(item, key)})).reduce((result,item) => Object.assign(result,item),{})
      }
      this.$router.push(router)
    },
    showCustomTh() {
      this.$refs.updateList.showCustomTh()
    },
    // 刷新列表
    refreshTable() {
      this.initSelected()
      this.$parent.$parent.initData(this.pageConfig.CRUD, this.pageConfig)
    },
    // 列表列数量改变时，重新计算列表展开位置长度
    changeTdNumber() {
      this.tdNumber = this.gettdsLength()
    },
    // 列复制功能相应函数,传入列序号，整理输入数据至剪切板
    doCopy(no, eleConfig) {
      const clipboardMsg = this.dataProcess(no, eleConfig)
      this.$copyText(clipboardMsg).then(function () {
      }, function () {
      })
      this.$Message.success('列数据已复制至剪切板！')
    },
    dataProcess(no, eleConfig) {
      const clipboardMsg = []
      for (const item of this.table.tableData) {
        clipboardMsg.push(this.renderValue(item, eleConfig))
      }
      return clipboardMsg.join('\n')
    },
    switchPage(router_path) {
      this.$router.push({
        name: router_path
      })
    },
    gettdsLength(){
      let count = 0
      if (this.table.selection){
        count++
      }
      if (this.table.btn){
        count++
      }
      this.table.tableEle.forEach(item => {
        if (item.display === true) {
          count++
        }
      })
      const config = this.table.tableEle
      for (let i=0; i<config.length; i++){
        if (config[i].display){
          this.firstShow = i
          break
        }
      }
      return count
    },
    loadDetail(item, index) {
      this.currentActive = this.currentActive === index ? -1 : index
      this.$root.$store.commit('changeTableExtendActive',this.currentActive)
      if (this.currentActive === -1){
        return
      }
      const func = this.pageConfig.table.isExtend.func
      this.$parent.$parent[func](item,index)
    },
    // 控制编辑按钮是否显示
    isShowEdit(rowNumber, columnNumber) {
      if (this.$root.$validate.isEmpty(this.activeEditInputNo)) {
        return
      }
      // 在设置可编辑字段(editable),并存在内容时显示编辑按钮
      // if (this.$root.$validate.isEmpty(val.editable) && this.$root.$validate.isEmpty(value[val.value])) {
      this.showEditRowNumber = rowNumber
      this.showEditColumnNumber = columnNumber
    },
    isShowEditXX(tableEle, rowNumber, columnNumber) {
      if (this.$root.$validate.isEmpty(tableEle.editable) && this.showEditRowNumber === rowNumber && this.showEditColumnNumber === columnNumber) {
        return true
      }
      return false
    },
    // 点击编辑按钮时相应
    showEditInput(value,tableEle,index, i) {
      this.isShowEditInput = !this.isShowEditInput
      this.activeEditInputNo = [index, i]
      setTimeout(() => {
        document.getElementById('editInput').focus()
      },300)
      this.editContent = value[tableEle.value]
    },
    // 在线编辑框失去焦点
    inputBlur(tableEle,value) {
      this.isShowEditInput = false
      this.activeEditInputNo = []
      const param = {}
      param[tableEle.value] = this.editContent
      this.$parent.$parent[tableEle.editable](value.id, param)
    },
    // 过滤框搜索功能
    searchFiltersContent() {
      this.filterListsBySearch = this.filterLists
      this.filterListsBySearch = this.filterListsBySearch.filter(item => item.label.toUpperCase().includes(this.filterParam.toUpperCase()))
    },
    // 无数据时系新增跳转
    goToConfig(tip){
      this.$router.push({
        name: tip.url,
        params: tip
      })
    },
    // 是否显示省略号
    ellipsis(value,val){
      const valuelength = this.renderValue(value, val)?this.renderValue(value, val).length:0
      const maxlength = val.overflow?val.overflow:20 // 通过配置overflow自定义显示字数
      return valuelength > maxlength ? true:false
    }
  },
  components: {
    // updateLists
  }
}
</script>

<style>
  .ivu-poptip-body {
    padding: 0;
  }
  .ivu-icon-funnel {
    color: #bbbec4;
  }

  .ivu-table-sort,.ivu-poptip-rel i.on {
    color: #5384FF;
  }
  .popfilter{
    max-height: 200px;
    overflow: auto;
  }

  /*隐藏poptip箭头
    因列表加入keep-alive在poptip中跳离返回后，poptip仍显示，
    现使用v-if控制跳离时隐藏，但箭头无法控制，所以此处将其隐藏，曲线救国*/
  .ivu-poptip-popper  .ivu-poptip-arrow {
    display: none;
  }

</style>
<style lang="less" scoped>
  .tag-f {
    border: 1px solid @color-blue;
    padding: 2px;
    color: @color-blue;
    border-radius: 4px;
    margin: 2px 4px;
    line-height: 18px;
  }
  .multiFilter /deep/.ivu-checkbox-group-item {
    display: block;
    padding: 0 16px;
    margin: 0;
    line-height: 28px;
  }

  .ui-ul-list ul {
    max-height: 250px;
    text-align: center;
  }

  .btn {
    width: 65px;
    height: 32px;
    padding: 4px 0;
  }
  .btn-primary{
    background-color: @color-blue;
    border-color: #D4E1EA;
  }
  .table-template {
    position: relative;
    .ui-table-container {
      overflow-x: auto;
      /*min-height: 400px;*/
      overflow-y: hidden;
    }
    /*滚动条样式*/
    .ui-table-container::-webkit-scrollbar { /*滚动条整体样式*/
      height: 6px;
    }

    .ui-table-container::-webkit-scrollbar-thumb { /*滚动条里面小方块*/
      border-radius: 6px;
      /*-webkit-box-shadow: inset 0 0 5px rgba(0, 0, 0, 0.2);*/
      background: rgba(0, 0, 0, 0.1);
    }

    .ui-table-container::-webkit-scrollbar-track { /*滚动条里面轨道*/
      /*-webkit-box-shadow: inset 0 0 5px rgba(0, 0, 0, 0.2);*/
      border-radius: 0;
      /*background: rgba(0, 0, 0, 0.1);*/
      background: white;
    }

    .ui-table-content-th {
      position: relative;
      /*width: 120px;*/
      height: 40px;
      padding: 2px 0px 2px 8px;
      font-size: 13px;
      color: #333333;
      letter-spacing: 0.7px;
    }
  }

  .th-border-bottom {
    border-bottom-width: 1px;
  }
  /*列表选择样式-开始 */
  .cart-item-check {
   padding-left: 4px;
  }

  .item-check-btn {
    border-radius: 3px;
    display: inline-block;
    width: 16px;
    height: 16px;
    border: 1px solid #ccc;
    text-align: center;
    vertical-align: middle;
    cursor: pointer;
    position: relative;
  }

  .item-check-btn .icon-ok {
    display: none;
    width: 100%;
    height: 100%;
    fill: #fff;
    transform: scale(0.8);
  }

  .item-check-btn.check {
    border-color: @color-blue;
  }
  .item-check-btn.check:after {
    content: "";
    position: absolute;
    left: 1px;
    bottom: 6px;
    width: 12px;
    height: 7px;
    border: 2px solid @color-blue;;
    border-top-color: transparent;
    border-right-color: transparent;
    transform: rotate(-45deg);
  }

  .item-check-btn.check .icon-ok {
    display: inline-block;
  }
  /*列表选择样式-结束 */

  .table{
    margin-bottom: 0;
  }
  .table th {
    // text-align: center;
    // padding: 0px 8px;
    padding-left: 8px;
    vertical-align: middle;
  }
  .tdoverflow{
    max-width: 20px;
    .cell{
      overflow: hidden;
    }
    /deep/ .ivu-tooltip-rel{
      display: block;
    }
    .poplableSty{
      // max-width: 150px;
      overflow: hidden;
      text-overflow: ellipsis;
      display: block;
      position: relative;
      white-space: nowrap;
      margin-bottom:0rem
    }
  }
  .table td {
    /*line-height: 10px;*/
    // max-width: 20px;
    font-size: 13px;
    color: #333333;
    letter-spacing: 0.7px;
    // text-align: center;
    vertical-align: middle;
    padding: 0 4px;
  }
  .td-center{
    // text-align: center;
  }

  .btn{
    margin-right: 4px;
    // width: 60px;
    // height: 32px;
  }

  .td-operation {
    button {
      margin-top: 3px;
      font-size: .2rem;
    }
  }

  .cell {
    padding: 0 4px;
    display: block;
    line-height: 40px;
    font-size: 12px;
    // overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    word-break: break-all;
    box-sizing: border-box;
    a {
      color: @color-blue;
      padding-left: 5px;
    }
    a:hover {
      color: @color-blue;
    }
  }

  .cell-icon {
    padding: 4px;
    color: #5384FF;
    cursor: pointer;
  }
  .cell-edit {
    // position: absolute;
    display: none;
    line-height: 36px;
    color: @color-blue;
    cursor: pointer;
  }
  .cell:hover{
    .cell-edit{
      display: inline-block;
    }
  }

  /*按钮组样式--开始*/
  .table-operation {
    position: absolute;
    top: -92px;
    right: 0;
    margin-left: 5px;
    text-align: center;
    height: 32px;
    /*width: 32px;*/
    color: #989898;
    /*font-size: 16px;*/
    cursor: pointer;
    /*border-radius: .2rem;*/
    /*background-color: RGB(97,137,186);*/
  }

  .table-operation-icon {
    // padding: 8px 10px;
    // border: 1px solid @color-gray-E;
    border-radius: 4px;
  }

  .label-name {
    text-align: left;
    margin-top: 8px;
    padding: 0 5px;
  }

  /*模态框样式-结束*/
  .batch-operation {
    display: inline-flex;
    .btn-group {
      margin-right: 5px;
    }
  }

  button:focus {
    box-shadow: none;
  }
  // .table tbody tr:nth-child(odd) {
  //   // background-color:  #F6F9FB;
  //   // background-color: #FFF;
  // }
  // .table tbody tr:nth-child(odd):hover{
  //   background-color: rgb(235, 237, 238);
  // }
  // .table tbody tr:nth-child(even):hover{
  //   background-color: rgb(235, 237, 238);
  // }
  .btn-operation {
    font-size: 12px;
    color: @color-blue;
    line-height: 20px;
    cursor: pointer;
  }
  .btn-pipe {
    color: #D8D8D8;
  }
  .Newline {
    white-space: normal;
    word-wrap: break-word;
  }

  // 表头筛选内容样式-开始
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
  .active {
    color: @color-blue;
    // background-color: @color-gray-F;
  }
  // 表头筛选内容样式-结束

  .tooltip-clipboard {
    background-color: rgba(70,76,91,.9);
    padding: 2px;
    border-radius: 4px;
  }
  .tooltip-clipboard:hover {
    background-color: #0080FF;
    color: white;
    cursor: pointer;
  }

  .btn-switch{
    width: auto;
    padding: 4px 8px;
    margin-bottom: 5px;
    margin-right: 20px;
    background-color: @color-blue;
    border-color: @color-blue;
    font-size: 14px;
    color: #FFFFFF;
    letter-spacing: 0.7px;
  }
  .extend-shadow {
    padding-right: 4px;
  }
.extendStyle{
  display: inline-block;
  line-height: 36px;
  padding: 0 3px;
  margin-top: 3px;
  float: left;
}
.bgc{
  // background-color: #fafafa !important;
}
.active{
  .ivu-icon{
    transform: rotate(90deg);
  }
}
.handleSty{
  position: absolute;
  right: 0;
  background-color: white;
  box-shadow: -2px 0 6px -2px rgba(0,0,0,.2);
}
.poplableSty{
  overflow: hidden;
  text-overflow: ellipsis;
  display: inline-block;
  white-space: nowrap;
  margin-bottom: -0.98rem;
}
  /*在线编辑input样式-开始*/
  .input-style {
    height:32px;
    border: 1px solid #ced4da;
    border-radius: .25rem;
  }
  .input-style:focus {
    box-shadow: none;
    color: #495057;
    background-color: #fff;
    border-color: #80bdff;
    outline: 0;
  }
  /*在线编辑input样式-结束*/

  /*tip搜索样式--开始*/
  .tip-search {
    width: 145px;
    line-height: 20px;
    border: 1px solid #ced4da;
    background-color: white;
    border-radius: .45rem;
    padding-left: 5px;
  }
  .tip-search:focus {
    outline: none;
    border-color: #80bdff;
  }
  /*tip搜索样式--结束*/
  .nodataSty{
    border-bottom: solid 1px #dee2e6;
    td{
      text-align: center;
      padding: 10px;
    }
  }

  // tooltip底部操作按钮
  .tooltip-footer {
    line-height: 32px;
    width: 200px;
    border-top: 1px solid #E9E9E9;
    font-size: 14px;
    padding: 0 16px
  }

  .tooltip-footer-btn {
    margin: 0;
    width: 50%;
    cursor: pointer;
  }
  .tooltip-footer-btn:hover {
    color: @color-blue;
  }
</style>
