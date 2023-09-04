(require 'org)

(let ((org-file "results.org"))
  (find-file org-file)
  (org-table-recalculate-buffer-tables)
  (save-buffer))
