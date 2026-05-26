const HtmlWebPackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');
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
      minimizer: [new CssMinimizerPlugin(), new TerserPlugin()],
      usedExports: true,
      sideEffects: true,
    },
    module: {
      rules: [
        {
          test: /\.js$/,
          include: [/src/, /node_modules\/react-json-editor-ajrm/],
          use: [
            {
              loader: 'babel-loader',
            },
          ],
        },
        // Plain CSS from node_modules (no Sass, no CSS modules)
        {
          test: /\.css$/,
          include: [/node_modules/],
          use: [
            {
              loader: DEV ? 'style-loader' : MiniCssExtractPlugin.loader,
            },
            {
              loader: 'css-loader',
              options: {
                modules: false,
                sourceMap: DEV,
              },
            },
          ],
        },
        // Global SCSS entry (index.scss) — no CSS modules
        {
          test: /\.scss$/,
          include: [/index\.scss/],
          use: [
            {
              loader: DEV ? 'style-loader' : MiniCssExtractPlugin.loader,
            },
            {
              loader: 'css-loader',
              options: {
                modules: false,
                sourceMap: DEV,
                importLoaders: 2,
              },
            },
            {
              loader: 'sass-loader',
              options: {
                sourceMap: DEV,
              },
            },
          ],
        },
        // Component SCSS — CSS modules
        {
          test: /\.scss$/,
          exclude: [/index\.scss/, /node_modules/],
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
                sourceMap: DEV,
                importLoaders: 2,
              },
            },
            {
              loader: 'sass-loader',
              options: {
                sourceMap: DEV,
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
        // Images
        {
          test: /\.(jpe?g|ico|gif|png|xml|webmanifest)$/,
          type: 'asset/resource',
          generator: {
            filename: 'assets/[name][ext]',
          },
        },
        // SVG (used as both images and icon fonts)
        {
          test: /\.svg$/,
          type: 'asset/resource',
          generator: {
            filename: 'assets/[name][ext]',
          },
        },
        // Fonts
        {
          test: /\.(eot|ttf|woff|woff2)$/,
          type: 'asset/resource',
          generator: {
            filename: 'assets/[name][ext]',
          },
        },
      ],
    },
    devServer: {
      open: true,
      client: {
        overlay: true,
        webSocketURL: { protocol: 'wss', pathname: '/webpack-hmr' },
      },
      webSocketServer: {
        type: 'ws',
        options: { path: '/webpack-hmr' },
      },
      server: 'https',
      historyApiFallback: true,
      proxy: {
        // the ones configures in webserver.Init
        '/ws': {
          target: 'https://localhost:6443',
          secure: false,
          ws: true,
        },
        '/register': {
          target: 'https://localhost:6443',
          secure: false,
        },
        '/login': {
          target: 'https://localhost:6443',
          secure: false,
        },
        '/logout': {
          target: 'https://localhost:6443',
          secure: false,
        },
        '/cert': {
          target: 'https://localhost:6443',
          secure: false,
        },
      },
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
