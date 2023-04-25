<template>
  <img
    class="circle-img"
    :width="imgWidth"
    :height="imgHeight"
    :src="$attrs.data.icon"
    :style="{
      transform: transformStyle
    }"
  />
</template>

<script>
export default {
  name: 'circle-rotate',
  data() {
    return {
      L: 280
    }
  },
  computed: {
    imgWidth() {
      const { value, total } = this.$attrs.data
      return this.L * ((parseInt(value, 10) / parseInt(total, 10)) || 0)
    },
    imgHeight() {
      return this.imgWidth
    },
    transformStyle() {
      const { deg, tx, ty } = this.$attrs.data
      return `rotate(${deg}) translate(${tx * this.imgWidth}px, ${ty * this.imgHeight}px)`
    }
  }
}
</script>

<style scoped lang="less">
.circle-img {
  position: absolute;
}
</style>