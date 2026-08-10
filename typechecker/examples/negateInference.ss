class 
    {
    updateTabSources(tabs, annotedSrc)
        {
		if annotatedSrc is false
			return
		n = tabs.GetAllTabCount()
		if annotatedSrc.Size() isnt n or not Number?(n)
			return
		for (i = 0; i < n; i  += 1)
			tabs.GetControl(i).Editor.Set(annotatedSrc[n - 1 - i])
		}
   }
