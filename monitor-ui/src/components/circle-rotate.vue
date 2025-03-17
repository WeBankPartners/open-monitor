<template>
  <img
    class="circle-img"
    :width="imgWidth"
    :height="imgHeight"
    :src="$attrs.data.icon"
    :style="{
      transform: transformStyle
    }"
    @click="handleClick"
  />
</template>

<script>
export default {
  name: 'circle-rotate',
  data() {
    return {
      L: 240
    }
  },
  computed: {
    imgWidth() {
      const { value, total } = this.$attrs.data
      return Math.max(this.L * ((parseInt(value, 10) / parseInt(total, 10)) || 0), 30)
    },
    imgHeight() {
      return this.imgWidth
    },
    transformStyle() {
      const { deg, tx, ty } = this.$attrs.data
      return `rotate(${deg}) translate(${tx * this.imgWidth}px, ${ty * this.imgHeight}px)`
    }
  },
  methods: {
    handleClick() {
      const { key, value } = this.$attrs.data
      if (+value > 0) {
        this.$emit('onFilter', {
          key: 'priority',
          value: key
        })
      }
    }
  }
}
</script>

<style scoped lang="less">
.circle-img {
  cursor: pointer;
  position: absolute;
  z-index: 100;
}
</style>
