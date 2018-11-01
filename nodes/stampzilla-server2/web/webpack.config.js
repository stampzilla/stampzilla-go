const GoogleFontsPlugin = require("@beyonk/google-fonts-webpack-plugin")
const HtmlWebPackPlugin = require("html-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const OptimizeCSSAssetsPlugin = require("optimize-css-assets-webpack-plugin");
const UglifyJsPlugin = require('uglifyjs-webpack-plugin');
const path = require('path');
const webpack = require('webpack');

const DEV = process.env.NODE_ENV === 'development';

module.exports = {
  devtool: DEV ? 'cheap-module-eval-source-map' : 'source-map',
  output: {
    filename: 'assets/[name].js',
  },
  optimization: {
    minimizer: [
      new UglifyJsPlugin({
        cache: true,
        parallel: true,
        sourceMap: true // set to true if you want JS source maps
      }),
      new OptimizeCSSAssetsPlugin({})
    ],
    usedExports: true,
    sideEffects: true
  },
  //optimization: {
    //minimize: true,
    //minimizer: [
      //new UglifyJsPlugin()
    //],
  //},
  module: {
    rules: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: {
          loader: "babel-loader"
        }
      },
      {
        test: /\.s?css$/,
        use: [{
          loader: DEV ? 'style-loader' : MiniCssExtractPlugin.loader,
        }, {
          loader: "css-loader", 
          options: {
            sourceMap: true
          }
        }, {
          loader: "sass-loader",
          options: {
            sourceMap: true
          }
        }]
      },
      {
        test: /\.html$/,
        use: [
          {
            loader: "html-loader"
          }
        ]
      },
      {
        test: /\.(png|jpg|gif)$/,
        use: [
          {
            loader: 'url-loader',
            options: {
              limit: 5000
            }
          }
        ]
      },
      {
        test: /\.(eot|svg|ttf|woff|woff2)$/,
        use: [
          {
            loader: 'file-loader',
            options: {
              //useRelativePath: !process.env.NODE_ENV,
              //publicPath: (DEV && CDN_URL === '') ? '/' : '', // Remove the default root slash because we load images from our CDN
              outputPath: 'assets/',
              publicPath: '/assets',
            },
          },
        ],
      },
    ]
  },
  devServer: {
    overlay: true
  },
  plugins: [
    new HtmlWebPackPlugin({
      template: "./src/index.html",
      filename: "./index.html"
    }),
    new webpack.ProvidePlugin({
      $: 'jquery',
      jQuery: 'jquery'
    }),
    new MiniCssExtractPlugin({
      filename: "assets/[name].css",
      chunkFilename: "assets/[id].css"
    }),
    new GoogleFontsPlugin({
      fonts: [
        { family: "Source Sans Pro", variants: [ "300", "400", "600", "700", "300italic", "400italic", "600italic" ] }
      ],
      path: 'assets/',
      filename: 'assets/fonts.css',
    })
  ],
};
