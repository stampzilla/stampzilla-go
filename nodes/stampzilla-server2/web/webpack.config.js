const GoogleFontsPlugin = require('@beyonk/google-fonts-webpack-plugin');
const HtmlWebPackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const OptimizeCSSAssetsPlugin = require('optimize-css-assets-webpack-plugin');
const SWPrecacheWebpackPlugin = require('sw-precache-webpack-plugin');
const WebpackPwaManifest = require('webpack-pwa-manifest');
// const path = require('path');
const webpack = require('webpack');

const DEV = process.env.NODE_ENV === 'development';

module.exports = {
  devtool: DEV ? 'cheap-module-eval-source-map' : 'source-map',
  output: {
    filename: 'assets/[name].js',
    publicPath: '/',
  },
  optimization: {
    minimizer: [
      new OptimizeCSSAssetsPlugin({}),
    ],
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
        include: [
          /src/,
          /node_modules\/react-json-editor-ajrm/,
        ],
        use: {
          loader: 'babel-loader',
        },
      },
      {
        test: /\.s?css$/,
        use: [{
          loader: DEV ? 'style-loader' : MiniCssExtractPlugin.loader,
        }, {
          loader: 'css-loader',
          options: {
            sourceMap: true,
          },
        }, {
          loader: 'sass-loader',
          options: {
            sourceMap: true,
          },
        }],
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
        test: /\.(png|jpg|gif)$/,
        use: [
          {
            loader: 'url-loader',
            options: {
              limit: 5000,
            },
          },
        ],
      },
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
      filename: 'assets/[name].css',
      chunkFilename: 'assets/[id].css',
    }),
    new GoogleFontsPlugin({
      fonts: [
        { family: 'Source Sans Pro', variants: ['300', '400', '600', '700', '300italic', '400italic', '600italic'] },
      ],
      path: 'assets/',
      filename: 'assets/fonts.css',
    }),
    new SWPrecacheWebpackPlugin({
      cacheId: 'stampzilla-go',
      dontCacheBustUrlsMatching: /\.\w{8}\./,
      filename: 'service-worker.js',
      minify: true,
      staticFileGlobsIgnorePatterns: [/\.map$/, /manifest\.json$/, /ws$/],
    }),
    new WebpackPwaManifest({
      filename: 'assets/manifest.[hash].json',
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
