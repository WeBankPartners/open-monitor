<template>
  <div class="basic-container">
    <img class="bg" src="../assets/img/bgd_main_cube.png" />
    <img class="cube" width="640" height="640" src="../assets/img/the_cube.png" />
    <img class="radar" v-if="$attrs.total === 0" width="373" height="373" src="../assets/img/corePRunning.png" />
    <div class="radar-text perfect-text" v-if="$attrs.total === 0">PERFECT</div>
    <div class="radar-text running-text" v-if="$attrs.total === 0">RUNNING...</div>
  </div>
</template>

<script>
export default {
  name: "alarm-assets-basic",
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
  }

  .radar {
    position: absolute;
    animation: radar-beam 5s infinite;
    animation-timing-function: linear;
  }

  .radar-text {
    position: absolute;
    color: #C5DCFE;
    font-size: 72px;
    font-weight: 800;

    &.perfect-text {
      padding-right: 800px;
    }
    &.running-text {
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