<template>
  <div>
    <div class="detail-table-template" v-for="(itemConfig, indexx) in detailConfig" :key="indexx">
      <h4 v-if="itemConfig.title">{{itemConfig.title}}</h4>
      <div v-if="!itemConfig.data.length">暂无数据</div>
      <table :class="[itemConfig.isExtendF ? 'table styleEX' : 'table table-bordered']" v-if="itemConfig.data.length">
        <thead>
          <tr>
            <th class="c-dark-gray" :width='itemConfig.scales[i]' v-for="(item,i) in itemConfig.config" :key="i">{{$t(item.title)}}</th>
          </tr>
        </thead>
        <tbody class="c-dark-gray">
          <template v-for="(item,index) in itemConfig.data">
            <tr :key="index">
              <td v-for="(key,i) in itemConfig.config" :key="i">
                <div v-if='itemConfig.isExtend && i === 0' class="extendStyle">
                  <a v-show="(indexx + '-' + index) !== currentActive" @click="loadDetail(item,indexx,index)"><i class="ivu-icon ivu-icon-ios-arrow-forward"></i></a>
                  <a v-show="(indexx + '-' + index) === currentActive" @click="loadDetail(item,indexx,index)" class="active"><i class="ivu-icon ivu-icon-ios-arrow-forward"></i></a>
                </div>
                <div v-if='!key.btn'>
                  <div v-for="(obj,index) in item[key.value]" :key='index'  class='textAlign'>
                    <template v-if="Array.isArray(item[key.value])">
                      <label v-if='obj.key'>{{obj.key}}:</label>
                      <span v-if='!obj.tip' :class="[obj.key ? colorblue : colorred]" class="breakword">&nbsp;{{obj.value}}</span>
                      <Tooltip placement="top" v-if='obj.tip'>
                        <span :class="[obj.key ? colorblue : colorred]" class="breakword cursorPointer">&nbsp;{{obj.value}}</span>
                        <div slot="content">
                          <div class="Newline">{{obj.tip}}</div>
                        </div>
                      </Tooltip>
                    </template>
                  </div>
                  <div v-if="!Array.isArray(item[key.value])">
                    <!-- 显示自定义tip -->
                    <Tooltip :content="item[key.value]" class="cell" :transfer="true"  v-if="key.toolTips">
                      <a @click="shadow(item, key)" v-if="'shadow' in key"> {{item[key.value]}}</a>
                      <span v-if="!('shadow' in key)"> {{item[key.value]}}</span>
                      <div slot="content">
                        <div class="Newline"> {{item[key.value]}}</div>
                      </div>
                    </Tooltip>
                    <!-- 不显示省略情况下 -->
                    <div class="cell" v-if="!key.toolTips && !ellipsis(item[key.value])">
                      <a @click="shadow(item, key)" v-if="'shadow' in key">{{item[key.value]}}</a>
                      <span v-if="!('shadow' in key) && !key.isIcon"> {{item[key.value] || '-'}}</span>
                      <span v-if="!('shadow' in key) && key.isIcon" :class="[item[key.value] ? colorgreen : colorred]"><i class="fa fa-circle"></i></span>
                    </div>
                    <!-- 显示省略且未自己配置tip情况下，自动添加tip,tip内容为整体内容 -->
                    <Tooltip v-if="ellipsis(item[key.value]) && !key.toolTips" :content="item[key.value]" class="cell" :transfer="true" :delay=1000  placement='bottom-start'>
                      <a @click="shadow(item, key)" v-if="'shadow' in key"> {{item[key.value]}}</a>
                      <span v-if="!('shadow' in key)"> {{item[key.value]}}</span>
                      <div slot="content">
                        <div class="Newline"> {{item[key.value]}}</div>
                      </div>
                    </Tooltip>
                  </div>
                </div>
                <div v-if='key.btn' class="">
                  <div style="display:inline-block" v-for="(btn_val,btn_i) in key.btn" :key="btn_i">
                    <span :id="btn_val.btn_func"
                          v-if="!item.hide && btn_val.btn_name !== 'more' && !btn_val.render"
                          class="btn-operation"
                          :style="{color: btn_val.color}"
                          @click="goToOpe(btn_val,item,index,indexx)"
                    >
                      {{$t(btn_val.btn_name)}}
                    </span>
                    <span :id="btn_val.btn_func"
                          v-if="item.hide && btn_val.btn_name !== 'more' && !btn_val.render && !item.hide[btn_i]"
                          class="btn-operation"
                          :style="{color: btn_val.color}"
                          @click="goToOpe(btn_val,item,index,indexx)"
                    >
                      {{$t(btn_val.btn_name)}}
                    </span>
                    <span :id="btn_val.btn_func"
                          v-if="btn_val.btn_name !== 'more' && btn_val.render"
                          class="btn-operation"
                          :style="{color: btn_val.color}"
                          @click="goToOpe(btn_val,item,index,indexx)"
                    >
                      {{renderValue(item, btn_val)}}
                    </span>
                    <span class="btn-pipe" v-if="item.hide && (!item.hide[btn_i] && !item.hide[btn_i + 1] ) && (btn_i !== key.btn.length - 1)">
                      |
                    </span>
                    <span class="btn-pipe" v-if="!item.hide && (btn_i !== key.btn.length - 1)">
                      |
                    </span>
                    <Poptip content="" placement="bottom" trigger="hover">
                      <div class="batch-operation" v-if="btn_val.btn_name === 'more'">
                        <div class="btn-group" role="group" @mouseover="btnMore(index)">
                          <span class="btn-operation">
                            {{$t('m_table_more')}}
                            <Icon type="chevron-down"></Icon>
                          </span>
                        </div>
                      </div>
                      <div slot="content" class="ui-ul-list">
                        <ul v-if="showMoreNumber === index">
                          <template v-for="(btn, moreIndex) in btn_val.more" >
                            <li @click="goToOpe(btn,item,index,indexx)" :key="moreIndex" :style="{color: btn.color}">{{$t(btn_val.btn_name)}}</li>
                          </template>
                        </ul>
                      </div>
                    </Poptip>
                  </div>
                </div>
              </td>
            </tr>
            <tr v-if="(indexx + '-' + index) === currentActive" :key="index + 0.5" class='bgc'>
              <td :colspan='itemConfig.config.length'>
                <tdSlot :name='itemConfig.isExtend.slot'> <slot :name='itemConfig.isExtend.slot'></slot> </tdSlot>
                <!-- tdslot在main.js中定义 -->
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>
  </div>
</template>
<script>
export default {
  data() {
    return {
      colorblue: 'colorblue',
      colorred: 'colorred',
      colorgreen: 'colorgreen',
      showMoreNumber: null, // 控制更多按钮显示序号
      currentActive: -1
    }
  },
  props: ['detailConfig'],
  mounted() {

  },
  methods: {
    shadow(item, val) {
      this.$router.push({
        name: val.shadow.path,
        params: {item}
      })
    },
    renderValue(item, config) {
      const func=config.render
      if (func){
        return func(item)
      }
      const expr=config.value
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
    /* ******** 点击按钮 *****  ******** */
    goToOpe(btn, value,index,indexx) {
      const func = btn.btn_func
      const callback = this.$parent.$parent[func] ? this.$parent.$parent[func] : this.$parent.$parent.$parent.$parent[func]// 作为扩展
      if (btn.btn_name === '删除'){
        this.$deleting.deleteFunc({
          delTip: value[btn.delTip],
          friendlyReminder: btn.friendlyReminder,
          okCallback: () => {
            callback(value,index,indexx)
          }
        })
      } else {
        callback(value,index,indexx)
      }
    },
    // 控制显示当前more按钮选项
    btnMore(index) {
      this.showMoreNumber = index
    },
    loadDetail(item, indexx, index) {
      this.currentActive = this.currentActive === indexx+'-'+index ? -1 : indexx+'-'+index
      if (this.currentActive === -1){
        return
      }
      const func = this.detailConfig[indexx].isExtend.func
      this.$parent.$parent[func](item,indexx, index)
      // detail.vue文件中配置 多table,多slot
      // <detailsPage :detailPageConfig="detailPageConfig">
      //   <div slot='tableExtend-1' scope="detailPageConfig.detailConfig[0].isExtend.data">
      //     <div>this is tableExtend-1</div>
      //     <div>{{detailPageConfig.detailConfig[0].isExtend.data.name}}</div>
      //   </div>
      //    <div slot='tableExtend-2' scope="detailPageConfig.detailConfig[0].isExtend.data">
      //     <div>this is tableExtend-2</div>
      //     <div>{{detailPageConfig.detailConfig[0].isExtend.data.name}}</div>
      //   </div>
      // </detailsPage>

      // 配置
      // detailConfig:[{
      //     title: '磁盘列表',
      //     config:[
      //       {title: '磁盘名', value: 'name', display: true},
      //       {title: '磁盘类型', value: 'volume_type', display: true},
      //       {title: '大小(GB)', value: 'size_gb', display: true},
      //       {title: '状态', value: 'status_state', display: true},
      //       {title: '创建时间', value: 'created_date', display: true},
      //     ],
      //     data:[],
      //     isExtend:{func:'getInfo',slot:'tableExtend-1',data:{}}, //重点！！！fun获取展开的数据
      //     scales:['25%','20%','15%','20%','20%']
      //   },]
    },
    // 是否显示省略号
    ellipsis(value){
      const valuelength = value?value.length:0
      const maxlength = 20
      return valuelength > maxlength ? true:false
    }
  },
  components: {}
}
</script>
<style lang="less" scoped>
.ui-ul-list ul {
    li{
      font-size: 13px;
      line-height: 30px;
      /*padding-left: 10px;*/
      color: #1b6499;
      /*margin-top: 10px;*/
      cursor: pointer;
    }
    li:hover{
      background-color: #f5f5f5;
      background: #DCEEFC;
    }
  }
.detail-table-template{
  font-size: 12px;
  padding-top: 10px;
  padding: 10px;
  font-weight: 400;
  /*margin: 20px 0px;*/
  background-color: #f2f2f2;
  margin-left: 24px;
}
.colorblue {
  color: @color-blue;
}
.colorred {
  color: red;
}
.colorgreen {
  color: #4ad64a;
}
.table thead {
  background: #fafafa;
}
.table {
  margin-bottom: 10px;
}
.table th {
  text-align: left;
  vertical-align: middle;
}
.table td {
  /*line-height: 10px;*/
  max-width: 20px;
  font-size: 12px;
  color: #333333;
  letter-spacing: 0.7px;
  text-align: left;
  vertical-align: middle;
  padding: 2px 6px;
}
.td-center {
  text-align: center;
}
h4 {
  text-align: left;
  font-size: 12px;
  font-weight: 550;
  margin-bottom: 10px;
  border-left: 5px solid@color-blue;
  padding-left: 4px;
}
.cell {
  display: block;
  line-height: 50px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  word-break: break-all;
  box-sizing: border-box;
  padding: 1px 0;
  a {
    color: @color-blue;
  }
  a:hover {
    color: #0092ee;
  }
}
.cell /deep/ .ivu-tooltip-rel {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
}
.Newline {
  white-space: normal;
  word-wrap: break-word;
}
.textAlign {
  text-align: left;
}
.breakword {
  word-wrap: break-word;
}

.cursorPointer {
  cursor: pointer;
}
.extendStyle{
  display: inline-block;
  line-height: 50px;
  padding-left: 10px;
  float: left;
}
.bgc{
  // background-color: #fafafa;
}
.active{
  .ivu-icon{
    transform: rotate(90deg);
  }
}
.extendBtn{
  display: inline-block;
  margin-left: -9%;
  margin-top: 17px;
  position: absolute;
}
.td-operation {
  text-align: left;
  button {
    margin-top: 3px;
    font-size: .2rem;
  }
}
.btn-operation {
    font-family: PingFangSC-Regular;
    font-size: 12px;
    color: @color-blue;
    line-height: 20px;
    /*padding-right: 4px;*/
    /*padding-left: 4px;*/
    cursor: pointer;
  }
  .td-operation {
    button {
      margin-top: 3px;
      font-size: .2rem;
    }
  }
  .btn-pipe {
    color: #D8D8D8;
  }
  .batch-operation {
    display: inline-flex;
    .btn-group {
      margin-right: 5px;
    }
  }
  .styleEX div .cell{
    line-height: 20px;
  }
  .styleEX .extendStyle{
    line-height: 20px;
  }
  .styleEX th{
    padding: 0.25rem 0;
  }
  .styleEX{
  thead {
    display: table;
    width: 100%;
    table-layout: fixed;
    th{
      width: auto;
      border-right-color: #fafafa;
    }
    th:last-child{
      border-right-color: #dee2e6;
    }
  }
  tbody {
    width: 100%;
    table-layout: fixed;
    display: block;
    max-height: 560px;
    overflow-y: scroll;
    tr {
      display: table;
      width: 100%;
      table-layout: fixed;
      // border-bottom: 1px solid #e9edf2;
      td{
        border-right-color: #fafafa;
      }
    }
  }
}

  /*.styleEX th, tr, td {*/
    /*border-width: 0;*/
  /*}*/
.styleEX th,td{
  border-bottom-width: 0;
}
.styleEX td{
  border-top-width: 0;
}
</style>
