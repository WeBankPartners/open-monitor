<template>
  <div class="login">
    <div class="login-con">
      <Card icon="log-in" title="欢迎登录" :bordered="false">
        <div class="form-con">
          <Form ref="loginForm" :model="form" :rules="rules" @keydown.enter.native="handleSubmit">
            <FormItem prop="userName">
              <Input v-model="form.userName" placeholder="请输入用户名">
                <span slot="prepend">
                  <Icon :size="16" type="ios-person"></Icon>
                </span>
              </Input>
            </FormItem>
            <FormItem prop="password">
              <Input type="password" v-model="form.password" placeholder="请输入密码">
                <span slot="prepend">
                  <Icon :size="14" type="md-lock"></Icon>
                </span>
              </Input>
            </FormItem>
            <FormItem>
              <Button @click="handleSubmit" type="primary" long>登录</Button>
            </FormItem>
          </Form>
        </div>
      </Card>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import {cookies} from '@/assets/js/cookieUtils'
export default {
  data () {
    return {
      form: {
        userName: 'test',
        ***REMOVED***
      },
      userNameRules: [{ required: true, message: '账号不能为空', trigger: 'blur' }],
      passwordRules: [{ required: true, message: '密码不能为空', trigger: 'blur' }],
    }
  },
  computed: {
    rules () {
      return {
        userName: this.userNameRules,
        password: this.passwordRules
      }
    }
  },
  methods: {
    handleSubmit () {
      let Base64 = require('js-base64').Base64
      axios({
        method: 'POST',
        url: 'http://129.204.99.160:38080/wecube-monitor/login',
        data: {
            userName: this.form.userName,
            password: Base64.encode(this.form.password)
        }
      }).then((response) => {
          if (response.status < 400) {
            cookies.setAuthorization(`${response.data.data.token}`)
            localStorage.username = response.data.user
            this.$Message.success('已登入!')
            this.$router.push({path: '/'})
          }
      })
      .catch((error) => {
        this.$Message.warning(this.$t('tips.failed'))
      });
    }
  }
}
</script>

<style lang="less" scope>
.login{
    width: 100%;
    height: 100%;
    background-image: url('../assets/img/login-bg.jpg');
    background-size: cover;
    background-position: center;
    position: relative;
    &-con{
        position: absolute;
        right: 39%;
        top: 50%;
        transform: translateY(-60%);
        width: 300px;
        &-header{
            font-size: 16px;
            font-weight: 300;
            text-align: center;
            padding: 30px 0;
        }
        .form-con{
            padding: 10px 0 0;
        }
        .login-tip{
            font-size: 10px;
            text-align: center;
            color: #c3c3c3;
        }
    }
}
</style>
