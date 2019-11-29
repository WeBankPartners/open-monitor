<template>
  <div class="index c-index">
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
  // width:1200px;
  min-height: 95vh;
  height: 100%;
  margin: 0 auto;
  padding: 10px;
}
</style>
