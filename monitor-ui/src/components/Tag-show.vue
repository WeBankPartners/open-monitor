<template>
  <div class="diy-tag" 
    :style="{color:colorList[choiceColor(tagName, index)] + ' !important',
            borderColor: colorList[choiceColor(tagName, index)] + ' !important'}">
    {{tagName}}
  </div>
</template>
<script>
import {colorList, endpointTag, randomColor} from '@/assets/config/common-config'
export default {
  data() {
    return {
      endpointTag: endpointTag,
      randomColor: randomColor,
      colorList: colorList,
      cacheColor: {}
    }
  },
  props:{
    index: {
      default: 0,
      type: Number
    },
    tagName: {
      default: '',
      type: String
    }
  },
  methods: {
    choiceColor (type,index) {
      if (endpointTag[type]) {
        return endpointTag[type]
      }
      let color = ''
      if (Object.keys(this.cacheColor).includes(type)) {
        color = this.cacheColor[type]
      } else {
        color = randomColor[index]
        this.cacheColor[type] = randomColor[index]
      }
      return color
    },
  }
}
</script>
<style lang="less">
  .diy-tag {
    width: 80px;
    text-align: center;
    display: inline-block;
    height: 24px;
    line-height: 22px;
    margin: 2px 4px 2px 0;
    padding: 0 8px;
    border: 1px solid #e8eaec;
    border-radius: 3px;
    // background: #f7f7f7;
    font-size: 12px;
    vertical-align: middle;
    opacity: 1;
    overflow: hidden;
  }
</style>