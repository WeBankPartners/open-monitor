<template>
  <Spin fix v-if="isFinalSpinShow" style="z-index: 1000000; opacity: 0.5">
    <Icon type="ios-loading" :size="25" class="spin-icon-load"></Icon>
    <div style="font-size: 20px">{{showText}}</div>
  </Spin>
</template>

<script>
export default {
  props: {
    isSpinShow: Boolean,
    showText: {
      type: String,
      default: ''
    }
  },
  watch: {
    isSpinShow: {
      handler(newVal) {
        this.isFinalSpinShow = newVal
      }
    }
  },
  data() {
    return {
      isFinalSpinShow: false
    }
  }
}
</script>

<style scoped>
.spin-icon-load {
  animation: ani-demo-spin 1s linear infinite;
}
@keyframes ani-demo-spin {
  from {
    transform: rotate(0deg);
  }
  50% {
    transform: rotate(180deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
