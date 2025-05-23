const path = require('path')
const vueConfig = require('./project-config/project-config.json')
module.exports = {
  devServer: {
    // hot: true,
    // inline: true,
		host: 'localhost',
    open: true,
    port: 3000,
    proxy: {
      '/': {
        target: process.env.BASE_URL,
				pathRewrite: {
						'^/': ''  // rewrite path
				}
      }
    }
  },
	assetsDir: process.env.PLUGIN === 'plugin'? '':'monitor',
	outputDir: process.env.PLUGIN === 'plugin'? 'plugin':'dist',
	productionSourceMap: false,
	// productionSourceMap: process.env.PLUGIN === 'plugin',
	chainWebpack: config => {
		config.when(process.env.PLUGIN === "plugin", config => {
      config
        .entry("app")
        .clear()
        .add(vueConfig.MAIN_PLUGIN_PATH); //作为插件时
    });
    config.when(!process.env.PLUGIN, config => {
      config
        .entry("app")
				.clear()
				.add(vueConfig.MAIN_PATH); //独立运行时
    })
		config.module.rule("images").test(/\.(png|jpeg|jpg|svg)$/).use("url-loader").loader("url-loader").options({
			limit: 1024 * 512
		})
    config.module.rule('worker').test(/\.worker\.js$/).use('worker-loader').loader('worker-loader')
      .options({ inline: 'fallback' }) // 内联模式，兼容旧浏览器
      .end();
    config.module.rule('js').exclude.add(/\.worker\.js$/);
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
	configureWebpack: {
    optimization: {
      runtimeChunk: 'single',
      splitChunks: {
        chunks: 'all',
        minSize: 20000, // 允许新拆出 chunk 的最小体积
        maxSize: 500000, // 设置chunk的最大体积为500KB
        automaticNameDelimiter: '-',
        cacheGroups: {
          defaultVendors: {
            test: /[\\/]node_modules[\\/]/,
            priority: -10
          },
          default: {
            minChunks: 2,
            priority: -20,
            reuseExistingChunk: true
          }
        }
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
	// configureWebpack: config => {
  //   if (process.env.PLUGIN === "plugin") {	
  //     config.optimization.splitChunks = {}
	// 		// config.optimization.minimize = false
	// 	}
  // },
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
