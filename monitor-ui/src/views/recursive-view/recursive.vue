<template>
  <div>
    <ul>
      <li v-for="(item, itemIndex) in recursiveViewConfig" class="tree-border" :key="itemIndex">
        <div @click="hide(itemIndex)" class="tree-title" :style="stylePadding">
          <span >
            <strong>| {{item.display_name}}</strong>
          </span>
        </div>
        <transition name="fade">
          <div v-show="item._isShow"  style="text-align:left">
            <recursive
              :increment="count"
              :params="params"
              v-if="item.children"
              :recursiveViewConfig="item.children"
            ></recursive>
            <div class="box xxx">
              <template v-for="(type, typeIndex) in monitorTypes">
                <Divider :key="type.type + typeIndex">{{type.type}}</Divider>
                <template v-for="(chartInfo,chartIndex) in item.charts">
                  <div :key="chartIndex + type" v-if="chartInfo.monitor_type === type.type" class="list">
                    <SingleChart
                      :chartInfo="chartInfo"
                      :chartIndex="chartIndex"
                      :params="params"
                      :refreshNow="refreshNow"
                      @editTitle="editTitle"
                      @sendConfig="receiveConfig"
                      @editShowLines="handleEditShowLines"
                    > </SingleChart>
                  </div>
                </template>
                <div v-for="(ep, epIndex) in type.empty" class="list" :key="ep + epIndex + Math.random()"></div>
              </template>
            </div>
          </div>
        </transition>
      </li>
    </ul>

    <ChartLinesModal
      :isLineSelectModalShow="isLineSelectModalShow"
      :chartId="setChartConfigId"
      @modalClose="onLineSelectChangeCancel"
    >
    </ChartLinesModal>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>

<script>
import {isEmpty} from 'lodash'
import SingleChart from '@/components/single-chart'
import ChartLinesModal from '@/components/chart-lines-modal'
export default {
  name: 'recursive',
  data() {
    return {
      inject: ['cacheColor'],
      modelConfig: {
        modalId: 'edit_Modal',
        modalTitle: 'm_button_chart_editTitle',
        saveFunc: 'titleSave',
        isAdd: true,
        config: [
          {
            label: 'm_tableKey_name',
            value: 'name',
            placeholder: 'm_tips_inputRequired',
            v_validate: 'required:true|min:2|max:60',
            disabled: false,
            type: 'text'
          }
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          name: null
        },
      },
      isLineSelectModalShow: false,
      setChartConfigId: '',
      refreshNow: false
    }
  },
  props: {
    params: {
      type: Object
    },
    recursiveViewConfig: {
      type: Array
    },
    increment: {
      type: Number,
      default: 0
    }
  },
  computed: {
    monitorTypes() {
      const types = {}
      this.recursiveViewConfig.forEach(recursice => {
        if (Array.isArray(recursice.charts)) {
          recursice.charts.forEach(chart => {
            if (chart.monitor_type in types) {
              types[chart.monitor_type] = types[chart.monitor_type] + 1
            } else {
              types[chart.monitor_type] = 1
            }
          })
        }
      })
      const monitorTypes = []
      const keys = Object.keys(types)
      keys.forEach(key => {
        monitorTypes.push({
          type: key,
          empty: 6 - types[key]%6
        })
      })
      return monitorTypes
    },
    count() {
      let c = this.increment
      return ++c
    },
    stylePadding(){
      return {
        'padding-left': this.count * 16 + 'px'
      }
    }
  },
  created() {
    window['view-config-selected-line-data'] = {}
    this.recursiveViewConfig.map(_ => {
      _._isShow = true
      if (_.charts) {
        const len = _.charts.length
        if (!len) {
          return
        }
        const remainder = 6 - len%6
        if (remainder) {
          const phZone = []
          for (let i = 0; i < remainder; i++) {
            phZone.push(Math.random())
          }
          _.phZone = phZone
        }
      }
    })
  },
  methods: {
    receiveConfig(chartItem) {
      this.$root.$eventBus.$emit('clearSingleChartInterval')
      this.$root.$eventBus.$emit('callMaxChart', chartItem)
    },
    hide(index) {
      this.recursiveViewConfig[index]._isShow = !this.recursiveViewConfig[index]._isShow
      this.$set(this.recursiveViewConfig, index, this.recursiveViewConfig[index])
    },
    editTitle(config) {
      this.modelConfig.addRow.name = config.title
      this.editChartConfig = config
      this.$root.JQ('#edit_Modal').modal('show')
    },
    titleSave() {
      const params = {
        chart_id: this.editChartConfig.id,
        metric: this.editChartConfig.metric,
        name: this.modelConfig.addRow.name
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.editTitle.api, params, () => {
        this.$root.JQ('#edit_Modal').modal('hide')
        this.$root.$eventBus.$emit('refreshRecursive', '')
      })
    },
    onLineSelectChangeCancel() {
      this.isLineSelectModalShow = false
      this.refreshNow = !this.refreshNow
    },
    handleEditShowLines(config) {
      this.setChartConfigId = config.id + ''
      if (isEmpty(window['view-config-selected-line-data'][config.id])) {
        window['view-config-selected-line-data'][config.id + ''] = {}
        config.legend.forEach(one => {
          window['view-config-selected-line-data'][config.id][one] = true
        })
      }
      this.isLineSelectModalShow = true
    }
  },
  components: {
    SingleChart,
    ChartLinesModal
  }
}
</script>

<style scoped lang="less">
  ul {
    padding: 0;
    margin: 0;
    list-style: none;
  }

  .tree-menu {
    height: 100%;
    padding: 0px 12px;
    border-right: 1px solid #e6e9f0;
  }

  .tree-menu-comm span {
    display: block;
    font-size: 12px;
    position: relative;
  }

  .tree-menu-comm span strong {
    display: block;
    width: 82%;
    position: relative;
    line-height: 22px;
    padding: 2px 0;
    padding-left: 5px;
    color: #161719;
    font-weight: normal;
  }

  .tree-title {
    margin-top: 1px;
    cursor: pointer;
    color: @blue-2;
  }
  .tree-border {
    border-top: 1px solid #9966;
    // border-right: none;
    // border-left: none;
    // border-top: none;
    padding: 4px 0;
    margin: 4px 0;
  }

  .box {
    display:flex;
    flex-wrap: wrap;
    justify-content: space-around;
  }
  .box .list{
    width: 580px;
  }
</style>
