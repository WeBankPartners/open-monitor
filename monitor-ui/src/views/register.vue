<template>
  <div class="login">
    <Row>
      <Col span="7" offset="9">
      <div class="login-con">
        <Card icon="log-in" title="注册" :bordered="false">
          <div class="form-con">
            <Form ref="loginForm"
                  :model="form"
                  :rules="rules"
                  @keydown.enter.native="handleSubmit"
                  :label-width="80"
            >
              <FormItem prop="username" label="用户名">
                <Input v-model="form.username" placeholder="">
                </Input>
              </FormItem>
              <FormItem prop="password" label="密码">
                <Input type="password" v-model="form.password" placeholder="">
                </Input>
              </FormItem>
              <FormItem prop="re_password" label="确认密码">
                <Input type="password" v-model="form.re_password" placeholder="">
                </Input>
              </FormItem>
              <FormItem prop="display_name" label="显示名">
                <Input v-model="form.display_name" placeholder="">
                </Input>
              </FormItem>
              <FormItem prop="email" label="邮箱">
                <Input v-model="form.email" placeholder="">
                </Input>
              </FormItem>
              <FormItem prop="phone" label="电话">
                <Input v-model="form.phone" placeholder="">
                </Input>
              </FormItem>
              <Button @click="handleSubmit" type="primary" long>注册</Button>
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
import { setToken} from '@/assets/js/cookies.ts'
export default {
  data() {
    return {
      form: {
        username: '',
        password: '',
        re_password: '',
        email: '',
        phone: '',
        display_name: ''
      },
      userNameRules: [{
        required: true,
        message: '用户名不能为空',
        trigger: 'blur'
      }],
      passwordRules: [{
        required: true,
        message: '密码不能为空',
        trigger: 'blur'
      }],
      rePasswordRules: [{
        required: true,
        message: '确认不能为空',
        trigger: 'blur'
      }],
      emailRules: [{
        required: true,
        message: '邮箱不能为空',
        trigger: 'blur'
      }],
      displayNameRules: [{
        required: true,
        message: '显示名不能为空',
        trigger: 'blur'
      }],
    }
  },
  computed: {
    rules() {
      return {
        username: this.userNameRules,
        password: this.passwordRules,
        re_password: this.rePasswordRules,
        email: this.emailRules,
        display_name: this.displayNameRules
      }
    }
  },
  methods: {
    handleSubmit() {
      const url = require('../../src/assets/js/baseURL').baseURL_config + '/monitor/register'
      this.$refs['loginForm'].validate(valid => {
        if (valid) {
          const Base64 = require('js-base64').Base64
          this.form.password = Base64.encode(this.form.password)
          this.form.re_password = Base64.encode(this.form.re_password)
          axios({
            method: 'POST',
            url,
            data: {
              ...this.form
            }
          }).then(response => {
            if (response.status < 400) {
              setToken(`${response.data.data.token}`)
              localStorage.username = response.data.data.user
              this.$Message.success(response.data.message)
              this.$router.push({path: '/'})
            }
          })
            .catch(() => {
              this.$Message.warning(this.$t('m_tips_failed'))
            })
        }
      })
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
    margin-top: 10%;
  }
  .form-con{
    padding: 10px 0 0;
  }
}
</style>
