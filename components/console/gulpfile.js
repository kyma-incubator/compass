const path = require('path');
const gulp = require('gulp');
const childProcess = require('child_process');
const log = require('fancy-log');
const clc = require('cli-color');
const { promisify } = require('util');

const exec = promisify(childProcess.exec);

process.on('unhandledRejection', err => {
  throw err;
});

// LIBRARIES
const libraries = [
  'common',
  'components/shared',
  'components/react',
  'components/generic-documentation',
  'shared',
];

// Installing libraries
libraries.forEach(lib => {
  gulp.task(`${lib}:install`, async () => {
    const packageName = path.resolve(__dirname, `./${lib}`);
    await install(packageName);
  });
});
gulp.task(
  'install:libraries',
  gulp.parallel(libraries.map(lib => `${lib}:install`)),
);

const install = async dir => {
  log.info(
    `Installing dependencies of ${clc.magenta(dir.replace(__dirname, ''))}`,
  );

  try {
    await exec(`npm install`, {
      cwd: dir,
    });
  } catch (err) {
    log.error(`Failed installing dependencies of ${dir}`);
    throw err;
  }
};

// CI libraries
libraries.forEach(lib => {
  gulp.task(`${lib}:ci`, async () => {
    const packageName = path.resolve(__dirname, `./${lib}`);
    await ci(packageName);
  });
});
gulp.task('ci:libraries', gulp.parallel(libraries.map(lib => `${lib}:ci`)));

const ci = async dir => {
  log.info(
    `Clean installing dependencies of ${clc.magenta(
      dir.replace(__dirname, ''),
    )}`,
  );

  try {
    await exec(`npm ci`, {
      cwd: dir,
    });
  } catch (err) {
    log.error(`Failed installing dependencies of ${dir}`);
    throw err;
  }
};

// Building libraries
libraries.forEach(lib => {
  gulp.task(`${lib}:build`, async () => {
    const packageName = path.resolve(__dirname, `./${lib}`);
    await build(packageName);
  });
});
gulp.task('build:libraries', gulp.series(libraries.map(lib => `${lib}:build`)));

const build = async dir => {
  log.info(`Building library ${clc.magenta(dir.replace(__dirname, ''))}`);

  try {
    await exec(`npm run build`, {
      cwd: dir,
    });
  } catch (err) {
    log.error(`Failed building library ${dir}`);
    throw err;
  }
};

// Watching libraries
gulp.task('watch:libraries', () => {
  libraries.forEach(lib => {
    gulp.watch([`./${lib}/src/**/*`], gulp.parallel(`${lib}:build`));
  });
});

// APPS
const apps = ['compass'];

// Installing apps
apps.forEach(app => {
  gulp.task(`${app}:install`, async () => {
    const packageName = path.resolve(__dirname, `./${app}`);
    await install(packageName);
  });
});
gulp.task('install:apps', gulp.parallel(apps.map(app => `${app}:install`)));
