
sell_bids = [
        {"price": 30, "quantity": 10, "linearlocation": 0},
        {"price": 30, "quantity": 10, "linearlocation": 10},
        {"price": 30, "quantity": 10, "linearlocation": 30},
        ]

buy_bids = [
        {"price": 40, "quantity": 10, "linearlocation": -3},
        {"price": 40, "quantity": 10, "linearlocation": -4},
        {"price": 40, "quantity": 10, "linearlocation": 24},
        ]

def multiplier(sell_bid, buy_bid):
    return 1/abs(sell_bid["linearlocation"] - buy_bid["linearlocation"])

def cost_function(sell_bid, buy_bid):
    return sell_bid["price"] - buy_bid["price"]

def hungarian_auction(sell_bids, buy_bids):
    
    def mount_matrix(sell_bids, buy_bids):
        matrix = []
        for sell_bid in sell_bids:
            row = []
            for buy_bid in buy_bids:
                row.append(multiplier(sell_bid, buy_bid) * cost_function(sell_bid, buy_bid))
            matrix.append(row)
        return matrix

    def subtract_min_from_rows(matrix):
        for row in matrix:
            min_value = min(row)
            for i in range(len(row)):
                row[i] -= min_value

    def subtract_min_from_columns(matrix):
        for i in range(len(matrix[0])):
            min_value = min([row[i] for row in matrix])
            for row in matrix:
                row[i] -= min_value

    def print_matrix(matrix):
        print("------------")
        for row in matrix:
            print(row)
        print("------------")
    
    matrix = mount_matrix(sell_bids, buy_bids)
    print_matrix(matrix)
    subtract_min_from_rows(matrix)
    print_matrix(matrix)
    subtract_min_from_columns(matrix)
    print_matrix(matrix)
# TODO: implement the rest, which requires counting the rows/columns with zeros and doing more things



hungarian_auction(sell_bids, buy_bids)
