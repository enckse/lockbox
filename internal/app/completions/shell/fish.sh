complete -c {{ $.Executable }} -f


function {{ $.Executable }}-completion
  set -f commands ""
  for line in ({{ $.ExportCommand }})
    set item (string split -m 1 "=" $line)
    set -f $item[1] $item[2]
  end
{{- range $idx, $value := $.Options }}
  {{- if gt $idx 0 }}
  set -f commands " $commands"
  {{ end }}
  if {{ $value.Conditional }}
    set -f commands "{{ $value.Key }}$commands"
  end
{{- end }}
  complete -c {{ $.Executable }} -n "not __fish_seen_subcommand_from $commands" -a "$commands"
  complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.HelpCommand }}; and test (count (commandline -opc)) -lt 3" -a "{{ $.HelpAdvancedCommand }} {{ $.HelpConfigCommand }}"
  if {{ $.Conditionals.Not.ReadOnly }}
    if {{ $.Conditionals.Not.AskMode }}
      complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.InsertCommand }} {{ $.MultiLineCommand }} {{ $.RemoveCommand }}; and test (count (commandline -opc)) -lt 3" -a "({{ $.DoList }})"
      complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.MoveCommand }}; and test (count (commandline -opc)) -lt 4" -a "({{ $.DoList }})"
    end
  end
  if {{ $.Conditionals.Not.CanTOTP }}
    set -f totpbase "{{ $.TOTPFindCommand }} {{ $.TOTPListCommand }}"
    set -f totps " $totpbase"
{{- range $idx, $value := $.TOTPSubCommands }}
  {{- if gt $idx 0 }}
    set -f totps " $totps"
  {{ end }}
    if {{ $value.Conditional }}
      set -f totps "{{ $value.Key }}$totps"
    end
{{- end }}
    complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.TOTPCommand }}; and not __fish_seen_subcommand_from $totps" -a "$totps"
    if {{ $.Conditionals.Not.AskMode }}
      complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.TOTPCommand }}; and __fish_seen_subcommand_from $totps; and not __fish_seen_subcommand_from $totpbase; and test (count (commandline -opc)) -lt 4" -a "({{ $.DoTOTPList }})"
    end
  end
  if {{ $.Conditionals.Not.CanClip }} 
    if {{ $.Conditionals.Not.AskMode }}
      complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.ClipCommand }}; and test (count (commandline -opc)) -lt 3" -a "({{ $.DoList}})"
    end
  end
  if {{ $.Conditionals.Not.AskMode }}
    complete -c {{ $.Executable }} -n "__fish_seen_subcommand_from {{ $.ShowCommand }} {{ $.JSONCommand }}; and test (count (commandline -opc)) -lt 3" -a "({{ $.DoList}})"
  end
end

{{ $.Executable }}-completion
