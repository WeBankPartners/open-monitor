<template>
  <div class="user-information-modify">
    <ul style=""> 
      <li>
        <label for="">{{$t('tableKey.name')}}：</label>
        <span>{{userInfo.name}}</span>
      </li>
      <li v-for="(info, infoKey) in infoConfig" :key="infoKey">
        <label for="">{{$t(info.label)}}：</label>
        <input v-if="activeKey === info.key" @blur="saveInfo(info.key)" v-model="userInfo[info.key]" type="text" class="form-control model-input">
        <span v-else>{{userInfo[info.key]}} <i @click="activeKey = info.key" class="fa fa-pencil-square-o" aria-hidden="true"></i></span>
      </li>
      <li>
        <label for="">{{$t('tableKey.activeDate')}}：</label>
        <span>{{userInfo.created}}</span>
      </li>
    </ul>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      infoConfig: [
        { label: 'tableKey.nickname', key: 'display_name' },
        { label: 'tableKey.email', key: 'email' },
        { label: 'tableKey.phone', key: 'phone' }
      ],
      activeKey: null,
      userInfo: {}
    }
  },
  mounted () {
    this.userInformation()
  },
  methods: {
    userInformation () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.setup.userInformation.get, {}, (responseData) => {
        this.userInfo = responseData
      })
    },
    saveInfo (key) {
      this.activeKey = null
      const params = {
        [key]: this.userInfo[key]
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.setup.userInformation.update, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.userInformation()
      })
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
.user-information-modify {
  min-width:500px;
  font-size: 15px;
  font-weight: 400;
  margin-top:20px;
  label {
    width:100px;
    text-align: right;
    margin-bottom: 0;
  }
  li {
    padding: 8px;
    i {
      margin-left: 20px;
      color: @color-blue;
    }
  }
}
.form-control {
  width: 20%;
}
.form-control:focus {
  box-shadow: none;
}
.model-input{
  font-size: 12px;
  display: inline-block;
}
</style>
