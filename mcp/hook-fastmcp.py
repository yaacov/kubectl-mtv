from PyInstaller.utils.hooks import collect_data_files, copy_metadata

# Collect all data files from fastmcp package
datas = collect_data_files('fastmcp') + copy_metadata('fastmcp')
