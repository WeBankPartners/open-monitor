<template>
  <div>
    <Modal v-model="DelConfig.isDisplay" width="360">
        <p slot="header" style="color:#f60;text-align:center">
            <Icon type="ios-information-circle"></Icon>
            <span>{{$t('delConfirm.title')}}</span>
        </p>
        <div style="text-align:center">
            <p>
              <span style="color:#2d8cf0"> {{DelConfig.displayConfig.name}} </span>
              {{$t('delConfirm.tip')}}
            </p>
        </div>
        <div slot="footer">
            <Button type="error" size="large" long @click="del">{{$t('delConfirm.button')}}</Button>
        </div>
    </Modal>
  </div>
</template>
<script>
  export default {
    data () {
      return {
        modal_loading: false,
      }
    },
    props:['DelConfig'],
    methods: {
      del () {
        this.DelConfig.callback()
        this.$Message.success(this.$t('tips.success'))
        this.DelConfig.isDisplay = false;
      }
    }
  }
</script>
