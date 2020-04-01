<template>
  <div class=" ">
    <ul>
      <li v-for="(item, itemIndex) in recursiveViewConfig" class="tree-border" :key="itemIndex">
        <div @click="hide(itemIndex)" class="tree-title" :style="stylePadding">
          <span >
            <strong>{{item.display_name}}</strong>
          </span>
        </div>
        <transition name="fade">
          <div v-show="item._isShow">
            <recursive
            :increment="count"
            :params="params"
            v-if="item.children"
            :recursiveViewConfig="item.children"></recursive>
            <div>
              <template v-for="(chartItemx,chartIndexx) in item.charts">
                <SingleChart 
                  :chartItemx="chartItemx" 
                  :chartIndex="chartIndexx" 
                  :key="chartIndexx" 
                  :params="params"
                  @sendConfig="receiveConfig"
                  > </SingleChart>
              </template>
            </div>
          </div>
        </transition>
      </li>
    </ul>
  </div>
</template>

<script>
import { EventBus } from "@/assets/js/event-bus.js"
import SingleChart from '@/components/single-chart'
export default {
  name: 'recursive',
  data() {
    return {
    }
  },
  props:{
    params: {
      type: Object
    },
    recursiveViewConfig:{
      type:Array
    },
    increment:{
      type:Number,
      default:0
    }
  },
  computed:{
    count () {
      var c = this.increment
      return ++c
    },
    stylePadding(){
      return {
        'padding-left':this.count * 16 + 'px'
      }
    }
  },
  created () {
    this.recursiveViewConfig.map((_) =>{
      _._isShow = true
    }) 
  },
  methods: {
    receiveConfig (chartItem) {
      EventBus.$emit("aaa", chartItem)
    },
    hide (index) {
      this.recursiveViewConfig[index]._isShow = !this.recursiveViewConfig[index]._isShow
      this.$set(this.recursiveViewConfig, index, this.recursiveViewConfig[index])
    }
  },
  components: {
    SingleChart
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
    border: 1px solid #9966;
    border-right: none;
    padding: 4px 0 4px 4px;
    margin: 4px 0 4px 4px;
  }
</style>
