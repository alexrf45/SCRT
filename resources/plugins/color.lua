return {
	"IlyasYOY/theme.nvim",
	dependencies = "tjdevries/colorbuddy.nvim",
	lazy = false,
	priority = 1000,
	config = function()
		vim.cmd.colorscheme("ilyasyoy")
	end,
}
