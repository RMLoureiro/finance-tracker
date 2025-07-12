class CreateTransactionTypes < ActiveRecord::Migration[8.0]
  def change
    create_table :transaction_types do |t|
      t.string :tittle

      t.timestamps
    end
  end
end
