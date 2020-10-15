const path = require('path');

module.exports = function(env) {
  return {
    entry: './index.js',
    output: {
      path: path.resolve(__dirname, 'build'),
      filename: 'index.js',
      library: '',
      libraryTarget: 'umd',
    },
    externals: {
      react: 'react',
      'react-dom': 'react-dom',
      'fundamental-react': 'fundamental-react',
    },
    module: {
      rules: [
        {
          test: /\.js$/,
          exclude: /node_modules/,
          loader: 'babel-loader',
          options: {
            presets: ['@babel/preset-env', '@babel/react'],
          },
        },
        {
          test: /\.s[ac]ss$/i,
          use: ['style-loader', 'css-loader', 'sass-loader'],
        },
        {
          test: /\.css$/i,
          use: ['style-loader', 'css-loader'],
        },
        {
          test: /\.(png|woff|woff2|eot|ttf|svg)$/,
          loader: 'url-loader?limit=100000',
        },
      ],
    },
    performance: {
      hints: false,
    },
    mode: env.production ? 'production' : 'development',
  };
};
