class AddMainToAccounts < ActiveRecord::Migration[8.0]
  def change
    add_column :accounts, :main, :boolean, default: false, null: false
  end
end 