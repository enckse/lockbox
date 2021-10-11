project('lockbox')
golang = find_program('go')
git_ver = run_command('git', 'log', '-n', '1', '--format=%h').stdout().strip()
vers = '-X main.version=' + git_ver
flags = ['-ldflags'] + [vers] + ['-trimpath', '-buildmode=pie', '-mod=readonly', '-modcacherw']

in_files = run_command('find', join_paths(meson.current_source_dir(), 'internal'), '-type', 'f', '-name', '*.go').stdout().strip().split()
in_files += 'go.mod'
in_files += 'go.sum'

progs = run_command('ls', join_paths(meson.current_source_dir(), 'cmd')).stdout().strip().split()

foreach p : progs
  p = custom_target(
      p,
      output: p,
      build_by_default: true,
      input: in_files,
      command: [ golang, 'build', flags, '-o','@OUTPUT@', join_paths(meson.current_source_dir(), 'cmd', p, 'main.go')],
  )
endforeach
test_dir = join_paths(meson.current_source_dir(), 'tests')
test('lb checks',
     find_program('bash'),
     workdir: test_dir,
     args: [join_paths(test_dir, 'run.sh'), meson.current_build_dir()])
