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
                  <div class="footer-style">测试账号:admin/admin</div>
                  <div class="register footer-style" @click="register">注册</div>
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
import { setToken } from '@/assets/js/cookies.ts'
export default {
  data () {
    return {
      form: {
        userName: '',
        password: ''
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
      const url =  require('../../src/assets/js/baseURL').baseURL_config + '/monitor/login'
      axios({
        method: 'POST',
        url: url,
        data: {
            userName: this.form.userName,
            password: Base64.encode(this.form.password)
        }
      }).then((response) => {
        if (response.status < 400) {
          setToken(`${response.data.data.token}`)
          localStorage.username = response.data.data.user
          this.$Message.success(response.data.message)
          this.$router.push({path: '/'})
        }
      })
      .catch(() => {
        this.$Message.warning(this.$t('m_tips_failed'))
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
    margin-bottom: 6px;
  }
  .register {
    float: right;
    cursor: pointer;
  }
  .register:hover {
    color: @blue-2;
  }
  .footer-style {
    line-height: 22px;
  }
}
</style>
