<template>
  <div class="basic-container">
    <img class="bg" src="../assets/img/bgd_main_cube.png" />
    <img class="cube" width="510" height="510" src="../assets/img/the_cube.png" />
    <template v-if="$attrs.noData">
      <img class="nodata" v-if="$attrs.total === 0" width="373" height="373" src="../assets/img/nodata.png" />
      <div class="nodata-text left" v-if="$attrs.total === 0">NO DATA</div>
      <div class="nodata-text right" v-if="$attrs.total === 0">RESPOND...</div>
    </template>
    <template v-else>
      <img class="radar" v-if="!$attrs.noData && $attrs.isRunning" width="373" height="373" src="../assets/img/corePRunning.png" />
      <div class="radar-text left" v-if="$attrs.total === 0">RUNNING</div>
      <div class="radar-text right" v-if="$attrs.total === 0">NORMALLY</div>
    </template>
  </div>
</template>

<script>
export default {
  name: 'alarm-assets-basic',
}
</script>

<style scoped lang="less">
.basic-container {
  display: flex;
  justify-content: center;
  align-items: center;

  .bg {
    position: absolute;
    top: 0;
    height: 100%;
    width: 100%;
  }

  .nodata {
    position: absolute;
  }

  .radar {
    position: absolute;
    animation: radar-beam 5s infinite;
    animation-timing-function: linear;
  }

  .nodata-text {
    position: absolute;
    color: #F2F3F7;
    font-size: 72px;
    font-weight: 800;

    &.left {
      padding-right: 800px;
    }
    &.right{
      padding-left: 900px;
    }
  }

  .radar-text {
    position: absolute;
    color: #C5DCFE;
    font-size: 72px;
    font-weight: 800;

    &.left {
      padding-right: 850px;
    }
    &.right{
      padding-left: 900px;
    }
  }
}

@keyframes radar-beam {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
</style>
