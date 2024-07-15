<template>
  <div class="index c-dark">
    <Menu></Menu>
    <div class="content">
      <router-view></router-view>
    </div>
  </div>
</template>

<script>
import Menu from '@/components/menu'
// import { watermark } from '@/assets/js/waterMark'
export default {
  name: 'index',
  data() {
    return {
    }
  },
  mounted() {
    // watermark({
    //   watermark_txt: 'Open-Monitor ' + localStorage.getItem('username'),
    //   watermark_fontsize: '14px',
    //   watermark_x_space: 300,
    //   watermark_y_space: 100,
    //   watermark_y: 100,
    //   watermark_alpha: 0.2
    // })
  },
  methods: {},
  components: {
    Menu
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style lang="less" scoped>
.index {
  min-width: 1200px;
  // background: white;
}
.content {
  height: calc(100vh - 210px);
  margin: 0 auto;
  padding: 10px;
}
</style>
