class CreateTransactions < ActiveRecord::Migration[8.0]
  def change
    create_table :transactions do |t|
      t.float :ammount
      t.references :account, null: false, foreign_key: true
      t.references :transaction_type, null: false, foreign_key: true

      t.timestamps
    end
  end
end
