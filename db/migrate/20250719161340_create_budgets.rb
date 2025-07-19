class CreateBudgets < ActiveRecord::Migration[8.0]
  def change
    create_table :budgets do |t|
      t.string :name
      t.float :amount
      t.date :start_date
      t.boolean :active, default: true
      t.belongs_to :account

      t.timestamps
    end
  end
end
