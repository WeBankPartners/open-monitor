<template>
  <Tooltip transfer :content="tagName" placement="top-start">
    <div class="diy-tag"
         :style="{
           color: getGroupColor(tagName) + ' !important',
           borderColor: getGroupColor(tagName) + ' !important'
         }"
    >
      {{tagName}}
    </div>
  </Tooltip>
</template>
<script>
import {endpointTag} from '@/assets/config/common-config'
export default {
  name: 'TagShow',
  data() {
    return {
      endpointTag,
      cacheColor: {}
    }
  },
  props: {
    index: {
      default: 0,
      type: Number
    },
    tagName: {
      default: '',
      type: String
    },
    list: {
      type: Array,
      default: () => []
    },
    name: {
      type: String,
      default: ''
    }
  },
  methods: {
    stringToColor(str) {
      let hash = 0
      for (let i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash)
      }
      let color = '#'
      for (let i = 0; i < 3; i++) {
        let value = (hash >> (i * 8)) & 0xFF
        // Add some offset to make the color darker
        value = Math.min(value + 50, 255)
        color += ('00' + value.toString(16)).substr(-2)
      }

      // Convert hex color to RGB
      const r = parseInt(color.substr(1, 2), 16)
      const g = parseInt(color.substr(3, 2), 16)
      const b = parseInt(color.substr(5, 2), 16)

      // Calculate brightness and saturation
      const y = 0.299 * r + 0.587 * g + 0.114 * b
      const s = 1 - 3 * Math.min(r, g, b) / (r + g + b)

      // If the brightness is too high or the saturation is too low, generate a new color
      if (y > 150 || s < 0.2) {
        return this.stringToColor(str + 'a')
      }

      return color
    },
    getGroupColor(str) {
      return this.stringToColor(str)
    }
  }
}
</script>
<style lang="less">
  .diy-tag {
    width: fit-content;
    max-width: 100px;
    text-overflow: ellipsis;
    overflow: hidden;
    white-space: nowrap;
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
  }
</style>
