const GoogleFontsPlugin = require('@beyonk/google-fonts-webpack-plugin');
const HtmlWebPackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const OptimizeCSSAssetsPlugin = require('optimize-css-assets-webpack-plugin');
const WebpackPwaManifest = require('webpack-pwa-manifest');
const TerserPlugin = require('terser-webpack-plugin');
// const path = require('path');
const webpack = require('webpack');

module.exports = (env, argv) => {
  const DEV = argv.mode !== 'production';
  return {
    devtool: DEV ? 'eval-cheap-source-map' : false,
    output: {
      filename: DEV ? 'assets/[name].js' : 'assets/[name].[contenthash].js',
      publicPath: '/',
    },
    optimization: {
      minimizer: [new OptimizeCSSAssetsPlugin({}), new TerserPlugin()],
      usedExports: true,
      sideEffects: true,
    },
    module: {
      rules: [
        {
          test: /\.js$/,
          // exclude: /node_modules\/(?![react-json-editor-ajrm])/,
          // exclude: /node_modules\/(?!(react-json-editor-ajrm)\/).*/,
          // exclude: /node_modules\/(?![react\-json\-editor\-ajrm])/,
          // exclude: /node_modules/,
          include: [/src/, /node_modules\/react-json-editor-ajrm/],
          use: [
            {
              loader: 'babel-loader',
            },
          ],
        },
        {
          test: /\.s?css$/,
          include: [/index\.scss/],
          use: [
            {
              loader: DEV ? 'style-loader' : MiniCssExtractPlugin.loader,
            },
            {
              loader: 'css-loader',
              options: {
                modules: false,
                sourceMap: true,
                importLoaders: 2,
              },
            },
            {
              loader: 'sass-loader',
              options: {
                sourceMap: true,
              },
            },
          ],
        },
        {
          test: /\.s?css$/,
          exclude: [/index\.scss/],
          use: [
            {
              loader: DEV ? 'style-loader' : MiniCssExtractPlugin.loader,
            },
            {
              loader: 'css-loader',
              options: {
                modules: {
                  mode: 'local',
                  localIdentName: '[name]__[local]___[hash:base64:5]',
                },
                sourceMap: true,
                importLoaders: 2,
              },
            },
            {
              loader: 'sass-loader',
              options: {
                sourceMap: true,
              },
            },
          ],
        },
        {
          test: /\.html$/,
          use: [
            {
              loader: 'html-loader',
            },
          ],
        },
        {
          test: /\.jpe?g$|\.ico$|\.gif$|\.png$|\.svg$|\.ico$|\.xml$|\.webmanifest$/,
          use: [
            {
              loader: 'file-loader',
              options: {
                name: '[name].[ext]',
                // useRelativePath: !process.env.NODE_ENV,
                // publicPath: (DEV && CDN_URL === '') ? '/' : '',
                // Remove the default root slash because we load images from our CDN
                outputPath: 'assets/',
                publicPath: '/assets',
              },
            },
          ],
        },
        // {
        // test: /\.(png|jpg|gif)$/,
        // use: [
        // {
        // loader: 'url-loader',
        // options: {
        // limit: 5000,
        // },
        // },
        // ],
        // },
        {
          test: /\.(eot|svg|ttf|woff|woff2)$/,
          use: [
            {
              loader: 'file-loader',
              options: {
                // useRelativePath: !process.env.NODE_ENV,
                // publicPath: (DEV && CDN_URL === '') ? '/' : '',
                // Remove the default root slash because we load images from our CDN
                outputPath: 'assets/',
                publicPath: '/assets',
              },
            },
          ],
        },
      ],
    },
    devServer: {
      overlay: true,
      historyApiFallback: true,
    },
    plugins: [
      new HtmlWebPackPlugin({
        template: './src/index.html',
        filename: './index.html',
      }),
      new webpack.ProvidePlugin({
        $: 'jquery',
        jQuery: 'jquery',
      }),
      new MiniCssExtractPlugin({
        filename: 'assets/[name].[contenthash].css',
        chunkFilename: 'assets/[id].[contenthash].css',
      }),
      new GoogleFontsPlugin({
        fonts: [
          {
            family: 'Roboto',
            variants: [
              '300',
              '400',
              '600',
              '700',
              '300italic',
              '400italic',
              '600italic',
            ],
          },
        ],
        path: '/',
        filename: 'assets/fonts.css',
      }),
      new WebpackPwaManifest({
        filename: 'assets/manifest.[contenthash].json',
        name: 'stampzilla-go',
        short_name: 'stampzilla',
        description: 'Homeautomation :)',
        background_color: '#01579b',
        theme_color: '#01579b',
        'theme-color': '#01579b',
        start_url: '/',
        icons: [
          // {
          // src: path.resolve('src/images/icon.png'),
          // sizes: [96, 128, 192, 256, 384, 512],
          // destination: path.join('assets', 'icons')
          // }
        ],
      }),
      new webpack.DefinePlugin({
        NODE_ENV: `${process.env.NODE_ENV}`,
      }),
    ],
  };
};
