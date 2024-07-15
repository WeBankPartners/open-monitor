<template>
  <div class="page-notice" :class="'page-notice-' + noticeConfig.type">
    <template v-for="(noticeItem, noticeIndex) in noticeConfig.contents">
      <p :key="noticeIndex">{{$t(noticeItem.tip)}}</p>
    </template>
  </div>
</template>
<script>
export default {
  props: ['noticeConfig']
}
</script>

<style  lang="less" scoped>
  // 页面提示信息样式--开始
  .page-notice {
    margin:16px 0;
    padding: 10px;
  }
  .page-notice-info {
    color: #5db558;
    background: #f4fbf5;
    border:1px solid #35b34a;
    border-radius: 4px;
  }
  .page-notice-warn {
    color: #b43537;
    background: #fcecec;
    border:1px solid #f6b5b5;
  }
  .page-notice-error {
  }
  // 页面提示信息样式--结束
</style>
