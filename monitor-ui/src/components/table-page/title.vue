<template>
  <div class="ibox-title">
    <h5>
      <!-- <i class="fa fa-reply-all fa-1x" v-if="titleConfig.backUrl" aria-hidden="true" @click="back"></i> -->
      <i class="backpage" v-if="titleConfig.backUrl" aria-hidden="true" @click="back"><img :src="backImg" alt="还回"/></i>
      <i class="fa-sitting" v-else></i>
      <span>{{titleConfig.name}}</span>
    </h5>
  </div>
</template>
<script>
// import backImg from '../../../../src/assets/back.svg'
// import backImg from '../../assets/back.svg'
export default {
  data() {
    return {
      backImg: ''
    }
  },
  props: ['titleConfig'],
  mounted() {
  },
  methods: {
    back() {
      this.$router.back(-1)
      // this.$router.push({
      //   path: this.titleConfig.backUrl,
      //   query: {
      //   }
      // })
    }
  },
  components: {}
}
</script>

<style lang="less" scoped>
  .backpage{
      display: inline-block;
      width: 30px;
      height: 30px;
      img{
        width: 19px!important;
        height: 15px!important;
        cursor: pointer;
      }
    }
  .ibox-title {
    padding-top: 20px;
    padding-bottom: 4px;
    border-width: 4px 0 0;
    /*padding-left: 32px;*/
    background-color: white;
    i{
      width:38px;
      color: #1ab394;
      padding: 8px 0px 0px 12px;
    }
    .fa-sitting {
      padding-left: 18px;
    }
    h5 {
      text-align: left;
      font-size: 16px;
      /*padding-left: 12px;*/
      color: #333333;
      letter-spacing: 0.8px;
    }
  }
</style>
