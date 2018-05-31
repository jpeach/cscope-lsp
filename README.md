# cscope-lsp

cscope line interface for a [Language Server](https://langserver.org).

## Vim Integration

To use `cscope-lsp` in your vim session, you need to do a bit of
configuration:

```vim
" Return the current cursor position as "file:line:col"
function! s:position()
    return expand('%') . ':' . line('.') . ':' . col('.')
endfunction

if has("cscope")

    :execute ':set cscopeprg=cscope-lsp'

    " The file argument to 'add' is ignored by cscope-lsp but
    " required by the vim cscope integration. Any file name
    " will do here.
    :execute ':cs add .gitignore'

    " cs: Find symbol
    map <Leader>cs :cs find s <C-R>=<SID>position()<CR><CR>

    " cg: Find definition
    map <Leader>cg :cs find g <C-R>=<SID>position()<CR><CR>

    " cc: Find callers
    map <Leader>cc :cs find c <C-R>=<SID>position()<CR><CR>

    " cd: Find callees
    map <Leader>cd :cs find d <C-R>=<SID>position()<CR><CR>

    " ct: Find text string
    map <Leader>ct :cs find t <C-R>=<SID>position()<CR><CR>

    " ce: Find egrep pattern
    map <Leader>ce :cs find e <C-R>=<SID>position()<CR><CR>

    " cf: Find file
    map <Leader>cf :cs find f <C-R>=<SID>position()<CR><CR>

    " ci: Find files #including this
    map <Leader>ci :cs find i <C-R>=<SID>position()<CR><CR>

endif
```
