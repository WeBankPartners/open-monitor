<template>
  <div class="index c-dark">
    <Menu></Menu>
    <div class="content">
      <router-view></router-view>
    </div>
    <Del :DelConfig="DelConfig"></Del>
  </div>
</template>

<script>
import Del from '@/components/del'
import Menu from "@/components/menu";
export default {
  name: "index",
  data() {
    return {
      DelConfig: {
        isDisplay:false,
        displayConfig: {
          name: ''
        },
        callback: null
      }
    }
  },
  methods: {
    delConfirm(config,callback){
      this.DelConfig.isDisplay = true
      this.DelConfig.displayConfig = config
      this.DelConfig.callback = callback
    }
  },
  components: {
    Menu,
    Del
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style lang="less" scoped>
.index {
  min-width: 1200px;
  // background: white;
}
.content {
  background-image: url("../assets/img/watermark-3.png");
  height: calc(100vh - 210px);
  margin: 0 auto;
  padding: 10px;
}
</style>
