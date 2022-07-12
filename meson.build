project('lockbox', version: 'development')
golang = find_program('go')
vers = '-X main.version=' + meson.project_version()
flags = ['-ldflags'] + [vers] + ['-trimpath', '-buildmode=pie', '-mod=readonly', '-modcacherw']

in_files = run_command('find', join_paths(meson.current_source_dir(), 'internal'), '-type', 'f', '-name', '*.go').stdout().strip().split()
in_files += 'go.mod'
in_files += 'go.sum'

progs = run_command('ls', join_paths(meson.current_source_dir(), 'cmd')).stdout().strip().split()

foreach p : progs
  cmd_file = join_paths(meson.current_source_dir(), 'cmd', p, 'main.go')
  target_files = in_files
  target_files += cmd_file
  p = custom_target(
      p,
      output: p,
      build_by_default: true,
      input: target_files,
      command: [ golang, 'build', flags, '-o','@OUTPUT@', cmd_file],
  )
endforeach
test_dir = join_paths(meson.current_source_dir(), 'tests')
test('lb checks',
     find_program('bash'),
     workdir: test_dir,
     args: [join_paths(test_dir, 'run.sh'), meson.current_build_dir()])
