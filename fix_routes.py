with open('blogo-server/internal/mods/blog/blog.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()

# Remove line 280 (0-indexed: 279) — duplicate projectTL.GET
# Remove line 288 (0-indexed: 287) — duplicate projectRes.GET
# After removing 280, the line numbers shift, so remove higher number first
del lines[287]  # projectRes.GET
del lines[279]  # projectTL.GET

with open('blogo-server/internal/mods/blog/blog.go', 'w', encoding='utf-8') as f:
    f.writelines(lines)
print('OK')
