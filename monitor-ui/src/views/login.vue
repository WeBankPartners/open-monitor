<template>
  <div class="login">
    <Row>
      <Col span="6" offset="9">
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
                <FormItem class="formItem-set">
                  <Button @click="handleSubmit" type="primary" long>登录</Button>
                </FormItem>
                <FormItem class="formItem-set">
                  <div class="register" @click="register">注册</div>
                </FormItem>
              </Form>
            </div>
          </Card>
        </div>
      </Col>
    </Row>
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
            localStorage.username = response.data.data.user
            this.$Message.success(response.data.msg)
            this.$router.push({path: '/'})
          }
      })
      .catch(() => {
        this.$Message.warning(this.$t('tips.failed'))
      });
    },
    register() {
      this.$router.push({path: '/register'})
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
  &-con {
    margin-top: 40%;
  }
  .form-con{
    padding: 10px 0 0;
  }
  .formItem-set {
    margin-bottom: 12px;
  }
  .register {
    float: right;
    cursor: pointer;
  }
  .register:hover {
    color: @blue-2;
  }
}
</style>
