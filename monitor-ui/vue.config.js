const path = require('path')

module.exports = {
	// devServer: {
	// 	proxy: {
  //     "wecube-monitor": {
  //       target: ""
  //     },
  //   }
	// },
	assetsDir: process.env.PLUGIN === 'plugin'? '':'wecube-monitor',
	productionSourceMap: process.env.PLUGIN !== 'plugin',
	chainWebpack: config => {
		config.when(process.env.PLUGIN === "plugin", config => {
      config
        .entry("app")
        .clear()
        .add("./src/main-plugin.js"); //作为插件时
    });
    config.when(!process.env.PLUGIN, config => {
      config
        .entry("app")
        .clear()
        .add("./src/main.js"); //独立运行时
    })
		const types = ['vue-modules', 'vue', 'normal-modules', 'normal']
		types.forEach(type => addStyleResource(config.module.rule('less').oneOf(type)))
	},
	css: {
		loaderOptions: {
			less: {
				javascriptEnabled: true
			}
		}
	},
	// configureWebpack: {
	// 	plugins: [
	// 		new webpack.ProvidePlugin({
	// 			$:"jquery",
	// 			jQuery:"jquery",
	// 			"windows.jQuery":"jquery"
	// 		})
	// 	]
	// },
	configureWebpack: config => {
    if (process.env.PLUGIN === "plugin") {
      config.optimization.splitChunks = {}
    }
  },
	pluginOptions: {
    pwa: {
      iconPaths: {
        favicon32: './favicon.ico',
        favicon16: './favicon.ico',
        appleTouchIcon: './favicon.ico',
        maskIcon: './favicon.ico',
        msTileImage: './favicon.ico'
      }
    }
  }
}
function addStyleResource(rule) {
	rule.use('style-resource')
	.loader('style-resources-loader')
	.options({
		patterns: [
			// 需要全局导入的less路径，
			path.resolve(__dirname, './src/assets/css/common.less')
		],
	})
}