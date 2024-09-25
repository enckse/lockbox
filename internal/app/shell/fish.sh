complete -c {{ $.Executable }} -f

{{- range $idx, $profile := $.Profiles }}

function {{ $profile.Name }}
  set -l commands {{ range $idx, $value := $profile.Options }}{{ if gt $idx 0}} {{ end }}{{ $value }}{{ end }}
  complete -c {{ $.Executable }} -n "not __fish_seen_subcommand_from $commands" -a "$commands"
  complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.HelpCommand }}; and test (count (commandline -opc)) -lt 3" -a "{{ $.HelpAdvancedCommand }}"
{{- if not $profile.ReadOnly }}
{{- if $profile.CanList }}
  complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.InsertCommand }} {{ $.MultiLineCommand }} {{ $.RemoveCommand }}; and test (count (commandline -opc)) -lt 3" -a "({{ $.DoList }})"
  complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.MoveCommand }}; and test (count (commandline -opc)) -lt 4" -a "({{ $.DoList }})"
{{- end}}
{{- end}}
{{- if $profile.CanTOTP }}
  set -l totps {{ $.TOTPListCommand }}{{ range $key, $value := .TOTPSubCommands }} {{ $value }}{{ end }}
  complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.TOTPCommand }}; and not __fish_seen_subcommand_from $totps" -a "$totps"
{{- if $profile.CanList }}
complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.TOTPCommand }}; and __fish_seen_subcommand_from $totps; and test (count (commandline -opc)) -lt 4" -a "({{ $.DoTOTPList }})"
{{- end}}
{{- end}}
{{- if $profile.CanList }}
  complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.ShowCommand }} {{ $.JSONCommand }}{{ if $profile.CanClip }} {{ $.ClipCommand }} {{end}}; and test (count (commandline -opc)) -lt 3" -a "({{ $.DoList}})"
{{- end}}
end
{{- end}}

function {{ $.Executable }}-completions
  {{ $.Shell }}
end

{{ $.Executable }}-completions
